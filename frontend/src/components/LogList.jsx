export default function LogList({ logs, onSelect, onDelete }) {
  if (!logs || logs.length === 0) {
    return (
      <div className="bg-white p-4 rounded border text-gray-500 italic">
        No webhooks received yet.
      </div>
    )
  }

  function handleDelete(id) {
    fetch(`/logs/${id}`, { method: "DELETE" })
      .then(() => onDelete?.())
      .catch(() => alert("Failed to delete webhook"))
  }

  return (
    <div className="bg-white p-4 rounded border">
      <h3 className="text-lg font-semibold mb-2">📬 Received Requests</h3>
      <ul className="space-y-2">
        {logs.map((log) => (
          <li
            key={log.id}
            className="flex justify-between items-center px-2 py-1 hover:bg-gray-100"
          >
            <button
              onClick={() => onSelect(log)}
              className="text-left w-full"
            >
              <div className="font-mono text-xs">{log.id.slice(0, 8)}</div>
              <div className="text-gray-500">{log.timestamp}</div>
            </button>
            <button
              onClick={() => handleDelete(log.id)}
              className="text-red-500 text-xs ml-2 hover:underline"
            >
              ❌
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
