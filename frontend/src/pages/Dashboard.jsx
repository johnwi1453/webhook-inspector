import { useEffect, useState } from "react"
import TokenStatus from "../components/TokenStatus"
import LogList from "../components/LogList"
import TestWebhookForm from "../components/TestWebhookForm"
import LogDetails from "../components/LogDetails"
import Header from "../components/Header"

export default function Dashboard() {
  const [status, setStatus] = useState(null)
  const [logs, setLogs] = useState([])
  const [selectedLog, setSelectedLog] = useState(null)
  const [error, setError] = useState(null)
  const [loginToastMsg, setLoginToastMsg] = useState(null)


const refreshLogs = () => {
  fetch("/logs")
    .then((res) => res.json())
    .then(setLogs)
    .catch(() => setLogs([]))
}

const refreshStatus = () => {
  fetch("/status")
    .then((res) => {
      if (!res.ok) throw new Error("Status not ready")
      return res.json()
    })
    .then(setStatus)
    .catch(() => {
      setStatus(null)
    })
}

useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const justLoggedIn = params.get("login") === "1"
    const justLoggedOut = params.get("logout") === "1"

    if (justLoggedIn) {
      setLoginToastMsg("✅ Logged in successfully")
    }

    if (justLoggedOut) {
      setLoginToastMsg("👋 Logged out successfully")
    }

    if (justLoggedIn || justLoggedOut) {
      setTimeout(() => setLoginToastMsg(null), 3000)

      const cleanUrl = window.location.origin + window.location.pathname
      window.history.replaceState({}, document.title, cleanUrl)
    }

    // fetch data
    fetch("/status")
      .then((res) => {
        if (!res.ok) throw new Error("Not authenticated or missing token")
        return res.json()
      })
      .then(setStatus)
      .catch((err) => setError(err.message))

    refreshLogs()
  }, [])



  return (
    <div className="min-h-screen bg-gray-50 text-sm text-gray-800">
      <Header />

      <div className="bg-blue-50 border border-blue-200 text-blue-800 p-4 rounded mb-4 max-w-3xl mx-auto">
      👋 <strong>Welcome to Webhook Inspector</strong> — Create tokens, inspect payloads, and test webhooks.
      </div>

      <div className="p-4 max-w-6xl mx-auto">
        {error && (
          <div className="bg-red-100 text-red-700 p-2 mb-4 rounded">{error}</div>
        )}

        {loginToastMsg && (
          <div className="bg-green-100 text-green-800 p-2 rounded mb-3">
            {loginToastMsg}
          </div>
        )}

        {status && <TokenStatus status={status} />}

        <div className="grid grid-cols-4 gap-4 mt-4">
          {/* LEFT: Log list */}
          <div className="col-span-1">
            <LogList logs={logs} onSelect={setSelectedLog} onDelete={refreshLogs} />
          </div>

          {/* MIDDLE: Test webhook form + status */}
          <div className="col-span-1">
            <TestWebhookForm token={status?.token} onSent={() => {
              refreshLogs()
              refreshStatus()
            }} />

          </div>

          {/* RIGHT: Selected log details */}
          <div className="col-span-2">
            <LogDetails log={selectedLog} onClose={() => setSelectedLog(null)} />
          </div>
        </div>

      </div>
    </div>
  )
}
