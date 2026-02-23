import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App'

// 全局错误处理
window.onerror = (message, source, lineno, colno, error) => {
  console.error('Global error:', message, source, lineno, colno, error)
  const errorDiv = document.createElement('div')
  errorDiv.style.cssText = 'position:fixed;top:0;left:0;right:0;background:#fee;padding:20px;z-index:9999;font-family:monospace;white-space:pre-wrap;color:#c00;border-bottom:2px solid #c00;'
  errorDiv.innerHTML = `<strong>错误:</strong> ${message}<br><small>${source}:${lineno}:${colno}</small>`
  document.body.appendChild(errorDiv)
}

// 捕获未处理的 Promise 拒绝
window.onunhandledrejection = (event) => {
  console.error('Unhandled rejection:', event.reason)
  const errorDiv = document.createElement('div')
  errorDiv.style.cssText = 'position:fixed;top:0;left:0;right:0;background:#fee;padding:20px;z-index:9999;font-family:monospace;white-space:pre-wrap;color:#c00;border-bottom:2px solid #c00;'
  errorDiv.innerHTML = `<strong>Promise 错误:</strong> ${event.reason}`
  document.body.appendChild(errorDiv)
}

const rootElement = document.getElementById('root')
if (!rootElement) {
  const errorDiv = document.createElement('div')
  errorDiv.style.cssText = 'position:fixed;top:0;left:0;right:0;background:#fee;padding:20px;z-index:9999;color:#c00;'
  errorDiv.innerHTML = '<strong>错误:</strong> 找不到 root 元素'
  document.body.appendChild(errorDiv)
  throw new Error('找不到 root 元素')
}

try {
  createRoot(rootElement).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
  console.log('React 应用已挂载')
} catch (error: any) {
  console.error('渲染错误:', error)
  const errorDiv = document.createElement('div')
  errorDiv.style.cssText = 'position:fixed;top:0;left:0;right:0;background:#fee;padding:20px;z-index:9999;color:#c00;'
  errorDiv.innerHTML = `<strong>渲染错误:</strong> ${error.message}<br><pre>${error.stack}</pre>`
  document.body.appendChild(errorDiv)
}