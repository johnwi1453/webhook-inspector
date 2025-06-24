export default function Login() {
  return (
    <div className="p-4">
      <h2 className="text-2xl font-semibold">🔐 Login</h2>
      <a
        className="mt-4 inline-block bg-black text-white px-4 py-2 rounded"
        href="http://localhost:8080/auth/github"
      >
        Login with GitHub
      </a>
    </div>
  )
}
