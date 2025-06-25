import { useState } from "react"

export default function TestWebhookForm({ token, onSent }) {
  const [jsonBody, setJsonBody] = useState(`{ "event": "ping", "message": "hello!" }`)
  const [sending, setSending] = useState(false)
  const [error, setError] = useState(null)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setSending(true)
    setError(null)

    try {
      const res = await fetch(`/api/hooks/${token}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: jsonBody,
      })

      if (!res.ok) throw new Error(await res.text())
      onSent() // let parent refresh logs
    } catch (err) {
      setError(err.message || "Failed to send")
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="bg-white p-4 rounded border">
      <h3 className="text-lg font-semibold mb-2">🧪 Send Test Webhook</h3>
      <form onSubmit={handleSubmit}>
        <textarea
          rows={6}
          value={jsonBody}
          onChange={(e) => setJsonBody(e.target.value)}
          className="w-full font-mono border rounded p-2 text-sm"
        />
        {error && <p className="text-red-600 mt-2">{error}</p>}
        <button
          type="submit"
          disabled={sending}
          className="mt-3 bg-black text-white px-4 py-2 rounded hover:bg-gray-800"
        >
          {sending ? "Sending..." : "Send Webhook"}
        </button>
      </form>
    </div>
  )
}
