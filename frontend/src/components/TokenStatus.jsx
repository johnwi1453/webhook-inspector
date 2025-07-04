export default function TokenStatus({ status }) {

  function handleReset() {
    if (!window.confirm("This will delete all webhook logs and assign a new token. Proceed?")) {
      return
    }

    fetch("/reset", { method: "POST" })
      .then((res) => {
        if (!res.ok) throw new Error("Reset failed")
        return res.json()
      })
      .then(() => {
        alert("✅ Token reset! You now have a new endpoint.")
        window.location.reload()
      })
      .catch(() => alert("❌ Failed to reset token"))
  }

  function handleNewToken() {
    if (!window.confirm("This will generate a brand new token and discard the old one. Proceed?")) return;

    fetch("/create")
      .then(() => {
        alert("🎯 New token created!")
        window.location.reload()
      })
      .catch(() => alert("❌ Failed to create token"))
  }

  const isPrivileged = status.privileged
  const baseUrl = window.location.origin
  const webhookUrl = `${baseUrl}/api/hooks/${status.token}`

  return (
    <div className="bg-white border rounded p-4 shadow-sm">
      <h3 className="text-lg font-semibold mb-2">🔐 Token Info</h3>

      <div className="mb-3 text-sm text-gray-600">
        This token is linked to your current session and used to receive webhooks at:
        <div className="mt-1 font-mono text-xs bg-gray-100 p-2 rounded">
          POST {webhookUrl}
        </div>
      </div>

      <ul className="space-y-1 text-sm">
        <li>
          <strong>Webhook Token:</strong> <code>{status.token}</code>
        </li>
        <li>
          <strong>Access Tier:</strong> {isPrivileged ? "Elevated (500 requests/day)" : "Basic (50 requests/day)"}
        </li>
        <li>
          <strong>Requests Sent:</strong> {status.requests_used}
        </li>
        <li>
          <strong>Requests Remaining:</strong> {status.requests_remaining}
        </li>
        <li>
          <strong>Token Expiry:</strong> {status.ttl}
        </li>
      </ul>

      <div className="flex space-x-2 mt-4">
        <button
          onClick={handleReset}
          className="px-4 py-1.5 rounded bg-red-100 hover:bg-red-200 text-red-800 text-sm font-medium transition"
        >
          🔄 Reset Token
        </button>

        <button
          onClick={handleNewToken}
          className="px-4 py-1.5 rounded bg-blue-100 hover:bg-blue-200 text-blue-800 text-sm font-medium transition"
        >
          🎯 New Token
        </button>
      </div>
    </div>
  )
}
