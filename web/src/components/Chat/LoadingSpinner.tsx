export default function LoadingSpinner() {
  return (
    <div className="flex items-center justify-center py-4">
      <div className="w-5 h-5 border-2 border-gray-300 border-t-emerald-500 rounded-full animate-spin" />
    </div>
  )
}
