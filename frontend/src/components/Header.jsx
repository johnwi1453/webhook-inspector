import { useEffect, useState } from "react"

export default function Header() {
  const [user, setUser] = useState(null)

  useEffect(() => {
    fetch("/me")
      .then((res) => res.json())
      .then((data) => {
        if (data.logged_in) setUser(data.username)
      })
      .catch(() => {})
  }, [])

  return (
    <header className="bg-black text-white px-6 py-3 flex justify-between items-center">
      <h1 className="text-lg font-semibold">Webhook Inspector</h1>

      <div className="text-sm">
        {user ? (
          <div className="flex items-center gap-4">
            <span>👤 {user}</span>
            <a
              href="/logout"
              className="bg-white text-black px-3 py-1 rounded hover:bg-gray-200"
            >
              Logout
            </a>
          </div>
        ) : (
          <a
            href="/auth/github"
            className="bg-white text-black px-3 py-1 rounded hover:bg-gray-200"
          >
            Login with GitHub
          </a>
        )}
      </div>
    </header>
  )
}
