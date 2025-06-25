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
                alert("✅ Token reset!")
                refreshStatus()
                refreshLogs()
            })
            .catch(() => alert("Failed to reset token"))

    }

  return (
    <div className="bg-white border rounded p-4 shadow-sm">
      <h3 className="text-lg font-semibold mb-2">🔐 Token Status</h3>
      <ul className="space-y-1">
        <li><strong>Token:</strong> <code>{status.token}</code></li>
        <li><strong>Privileged:</strong> {status.privileged ? "✅ Yes" : "❌ No"}</li>
        <li><strong>Used:</strong> {status.requests_used}</li>
        <li><strong>Remaining:</strong> {status.requests_remaining}</li>
        <li><strong>TTL:</strong> {status.ttl_seconds}s</li>
      </ul>
      <button
        onClick={handleReset}
        className="mt-3 bg-red-100 text-red-800 px-3 py-1 rounded hover:bg-red-200 text-sm"
        >
        🔄 Reset Token
        </button>

    </div>
    
  )
}
