export default function LogDetails({ log, onClose }) {
  if (!log) return <p className="text-gray-400 italic">Select a request to see details.</p>

  return (
    <div className="bg-white p-4 rounded border shadow-sm relative">
      <button
        className="absolute top-2 right-2 text-gray-500 hover:text-black"
        onClick={onClose}
        title="Close"
      >
        &times;
      </button>

      <h3 className="text-lg font-semibold mb-2">Webhook Details</h3>

      <p className="text-xs text-gray-500 mb-2">ID: {log.id}</p>
      <p className="text-xs text-gray-500 mb-2">Time: {new Date(log.timestamp).toLocaleString()}</p>

      <div className="mb-2">
        <strong>Method:</strong> {log.method}
      </div>

      <div className="mb-2">
        <strong>Headers:</strong>
        <pre className="bg-gray-100 p-2 text-xs overflow-x-auto">{JSON.stringify(log.headers, null, 2)}</pre>
      </div>

      <div>
        <strong>Body:</strong>
        <pre className="bg-gray-100 p-2 text-xs overflow-x-auto">{log.body}</pre>
      </div>
    </div>
  )
}
