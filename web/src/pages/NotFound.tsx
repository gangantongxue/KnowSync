import { Link } from 'react-router-dom'

export default function NotFound() {
  return (
    <div className="h-full flex flex-col items-center justify-center text-gray-400">
      <div className="text-5xl mb-4">404</div>
      <p className="text-sm mb-4">页面不存在</p>
      <Link to="/chat" className="text-sm text-emerald-600 hover:underline">回到首页</Link>
    </div>
  )
}
