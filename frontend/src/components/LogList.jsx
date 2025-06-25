export default function LogList({ logs, onSelect }) {
  if (!logs || logs.length === 0) {
    return (
      <div className="bg-white p-4 rounded border text-gray-500 italic">
        No webhooks received yet.
      </div>
    )
  }

  return (
    <div className="bg-white p-4 rounded border">
      <h3 className="text-lg font-semibold mb-2">📬 Received Requests</h3>
      <ul className="space-y-2">
        {logs.map((log) => (
          <li
            key={log.id}
            className="p-2 border rounded hover:bg-gray-100 cursor-pointer"
            onClick={() => onSelect(log)}
          >
            <div className="font-mono text-sm">{log.method}</div>
            <div className="text-xs text-gray-500">
              {new Date(log.timestamp).toLocaleString()}
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}
