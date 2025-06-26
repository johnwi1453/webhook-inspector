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
      .then((data) => {
        alert("✅ Token reset! You now have a new endpoint.")
        window.location.reload()
      })
      .catch(() => alert("❌ Failed to reset token"))
  }

  return (
    <div className="bg-white border rounded p-4 shadow-sm">
      <h3 className="text-lg font-semibold mb-2">🔐 Token Status</h3>
      <ul className="space-y-1">
        <li><strong>Token:</strong> <code>{status.token}</code></li>
        <li><strong>Privileged:</strong> {status.privileged ? "✅ Yes" : "❌ No"}</li>
        <li><strong>Used:</strong> {status.requests_used}</li>
        <li><strong>Remaining:</strong> {status.requests_remaining}</li>
          <li><strong>TTL:</strong> {status.ttl}s</li>
      </ul>
      <div className="flex space-x-2 mt-4">
        <button
          onClick={() => {
            if (!window.confirm("Clear all webhook logs for this token?")) return;

            fetch("/reset", { method: "POST" })
              .then((res) => {
                if (!res.ok) throw new Error()
                return res.json()
              })
              .then(() => {
                alert("✅ Logs cleared and token reset.")
                window.location.reload()
              })
              .catch(() => alert("❌ Failed to reset token"))
          }}
          className="px-4 py-1.5 rounded bg-red-100 hover:bg-red-200 text-red-800 text-sm font-medium transition"
        >
          🔄 Reset Token
        </button>

        <button
          onClick={() => {
            if (!window.confirm("This will generate a brand new token and delete the old one. Proceed?")) return;

            fetch("/create")
              .then(() => {
                alert("🎯 New token created!")
                window.location.reload()
              })
              .catch(() => alert("❌ Failed to create token"))
          }}
          className="px-4 py-1.5 rounded bg-blue-100 hover:bg-blue-200 text-blue-800 text-sm font-medium transition"
        >
          🎯 New Token
        </button>
      </div>



    </div>
    
  )
}
