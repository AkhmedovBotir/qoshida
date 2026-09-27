import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import './index.css'

const rootTag = document.getElementById('root')

if (!rootTag) {
  throw new Error('Root element topilmadi')
}

createRoot(rootTag).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
