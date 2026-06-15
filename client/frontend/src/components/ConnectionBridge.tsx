import { useEffect, useRef } from 'react'
import type { ServiceStatus } from '../lib/types'

interface ConnectionBridgeProps {
  status: ServiceStatus
}

export function ConnectionBridge({ status }: ConnectionBridgeProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const animRef = useRef<number>(0)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')!
    const displayW = canvas.offsetWidth
    const displayH = canvas.offsetHeight
    canvas.width = displayW * 2
    canvas.height = displayH * 2
    ctx.scale(2, 2)

    const lineY = displayH / 2
    const startX = 60
    const endX = displayW - 60
    const dotPos = { x: startX }

    function draw() {
      ctx.clearRect(0, 0, displayW, displayH)

      const lineColor = status === 'running' ? '#5856D6' : '#86868B'
      ctx.strokeStyle = lineColor
      ctx.lineWidth = 1.5
      ctx.beginPath()
      ctx.moveTo(startX, lineY)
      ctx.lineTo(endX, lineY)
      ctx.stroke()

      ctx.fillStyle = status === 'running' ? '#5856D6' : '#86868B'
      ctx.beginPath()
      ctx.arc(startX, lineY, 8, 0, Math.PI * 2)
      ctx.fill()
      ctx.beginPath()
      ctx.arc(endX, lineY, 8, 0, Math.PI * 2)
      ctx.fill()

      ctx.font = '12px -apple-system, sans-serif'
      ctx.fillStyle = status === 'running' ? '#1D1D1F' : '#86868B'
      ctx.textAlign = 'center'
      ctx.fillText('Agent', startX, lineY - 20)
      ctx.fillText('飞书', endX, lineY - 20)

      if (status === 'running') {
        dotPos.x += 0.5
        if (dotPos.x > endX) dotPos.x = startX

        ctx.fillStyle = '#5856D6'
        ctx.beginPath()
        ctx.arc(dotPos.x, lineY, 3, 0, Math.PI * 2)
        ctx.fill()

        ctx.fillStyle = 'rgba(88, 86, 214, 0.3)'
        ctx.beginPath()
        ctx.arc(dotPos.x - 6, lineY, 2, 0, Math.PI * 2)
        ctx.fill()
      }

      animRef.current = requestAnimationFrame(draw)
    }

    draw()
    return () => cancelAnimationFrame(animRef.current)
  }, [status])

  return <canvas ref={canvasRef} className="w-full h-20" />
}
