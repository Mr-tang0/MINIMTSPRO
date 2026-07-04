<template>
  <div class="roi-page">
    <header class="toolbar">
      <div>
        <div class="title">选取标记 {{ label }}</div>
        <div class="hint">左键拖拽绘制方框，滚轮缩放图片，确认后开始 DIC 追踪</div>
      </div>
      <div class="zoom-badge">{{ Math.round(scale * 100) }}%</div>
    </header>

    <main ref="stageRef" class="stage" @wheel.prevent="handleWheel">
      <canvas
        ref="canvasRef"
        class="canvas"
        @mousedown="handleMouseDown"
        @mousemove="handleMouseMove"
        @mouseup="handleMouseUp"
        @mouseleave="handleMouseUp"
      ></canvas>
      <div v-if="loading" class="overlay-text">正在加载最新相机帧...</div>
      <div v-else-if="!imageLoaded" class="overlay-text error">未获取到相机图像</div>
    </main>

    <footer class="footer">
      <div class="roi-info">
        <span>ROI</span>
        <strong v-if="roi">X {{ Math.round(roi.x) }} / Y {{ Math.round(roi.y) }} / W {{ Math.round(roi.width) }} / H {{ Math.round(roi.height) }}</strong>
        <strong v-else>请绘制方框</strong>
      </div>
      <div class="actions">
        <button class="btn secondary" @click="Window.Close()">×</button>
        <button class="btn primary" :disabled="!roi" @click="confirmROI">√</button>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Window } from '@wailsio/runtime'
import { GetLatestFrameForROI, SetROI } from '../../bindings/changeme/backend/hikcameraservice'

const route = useRoute()
const label = computed(() => route.query.label === 'B' ? 'B' : 'A')

const stageRef = ref(null)
const canvasRef = ref(null)
const loading = ref(true)
const imageLoaded = ref(false)
const roi = ref(null)

const img = new Image()
let imgWidth = 0
let imgHeight = 0
let scale = ref(1)
let offsetX = 0
let offsetY = 0
let drawing = false
let startPoint = null
let currentPoint = null

const canvasPointToImagePoint = (clientX, clientY) => {
  const rect = canvasRef.value.getBoundingClientRect()
  return {
    x: (clientX - rect.left - offsetX) / scale.value,
    y: (clientY - rect.top - offsetY) / scale.value
  }
}

const clampPoint = (p) => ({
  x: Math.max(0, Math.min(imgWidth, p.x)),
  y: Math.max(0, Math.min(imgHeight, p.y))
})

const resizeCanvas = () => {
  const canvas = canvasRef.value
  const stage = stageRef.value
  if (!canvas || !stage) return

  canvas.width = stage.clientWidth
  canvas.height = stage.clientHeight

  if (imageLoaded.value && scale.value === 1) {
    const fit = Math.min(canvas.width / imgWidth, canvas.height / imgHeight) * 0.92
    scale.value = Math.max(0.1, fit)
    offsetX = (canvas.width - imgWidth * scale.value) / 2
    offsetY = (canvas.height - imgHeight * scale.value) / 2
  }
  draw()
}

const draw = () => {
  const canvas = canvasRef.value
  if (!canvas) return

  const ctx = canvas.getContext('2d')
  ctx.clearRect(0, 0, canvas.width, canvas.height)
  ctx.fillStyle = '#020617'
  ctx.fillRect(0, 0, canvas.width, canvas.height)

  if (!imageLoaded.value) return

  ctx.save()
  ctx.translate(offsetX, offsetY)
  ctx.scale(scale.value, scale.value)
  ctx.drawImage(img, 0, 0)

  const drawRect = (rect, color) => {
    ctx.strokeStyle = color
    ctx.lineWidth = 2 / scale.value
    ctx.setLineDash([8 / scale.value, 4 / scale.value])
    ctx.strokeRect(rect.x, rect.y, rect.width, rect.height)
    ctx.setLineDash([])
    ctx.fillStyle = color
    ctx.font = `${18 / scale.value}px Inter, sans-serif`
    ctx.fillText(label.value, rect.x + rect.width + 6 / scale.value, rect.y + 18 / scale.value)
  }

  if (roi.value) drawRect(roi.value, '#22c55e')
  if (drawing && startPoint && currentPoint) {
    const x = Math.min(startPoint.x, currentPoint.x)
    const y = Math.min(startPoint.y, currentPoint.y)
    const width = Math.abs(currentPoint.x - startPoint.x)
    const height = Math.abs(currentPoint.y - startPoint.y)
    drawRect({ x, y, width, height }, '#38bdf8')
  }
  ctx.restore()
}

const handleWheel = (event) => {
  if (!imageLoaded.value) return

  const canvas = canvasRef.value
  const rect = canvas.getBoundingClientRect()
  const mouseX = event.clientX - rect.left
  const mouseY = event.clientY - rect.top
  const imageX = (mouseX - offsetX) / scale.value
  const imageY = (mouseY - offsetY) / scale.value

  const factor = event.deltaY < 0 ? 1.12 : 0.88
  const nextScale = Math.max(0.08, Math.min(8, scale.value * factor))
  scale.value = nextScale
  offsetX = mouseX - imageX * scale.value
  offsetY = mouseY - imageY * scale.value
  draw()
}

const handleMouseDown = (event) => {
  if (!imageLoaded.value || event.button !== 0) return
  drawing = true
  startPoint = clampPoint(canvasPointToImagePoint(event.clientX, event.clientY))
  currentPoint = startPoint
  draw()
}

const handleMouseMove = (event) => {
  if (!drawing) return
  currentPoint = clampPoint(canvasPointToImagePoint(event.clientX, event.clientY))
  draw()
}

const handleMouseUp = () => {
  if (!drawing || !startPoint || !currentPoint) return
  drawing = false

  const x = Math.min(startPoint.x, currentPoint.x)
  const y = Math.min(startPoint.y, currentPoint.y)
  const width = Math.abs(currentPoint.x - startPoint.x)
  const height = Math.abs(currentPoint.y - startPoint.y)
  if (width >= 8 && height >= 8) {
    roi.value = { x, y, width, height }
  }
  startPoint = null
  currentPoint = null
  draw()
}

const confirmROI = async () => {
  if (!roi.value) return
  try {
    await SetROI(label.value, roi.value)
    await Window.Close()
  } catch (err) {
    alert(`ROI 设置失败：${err}`)
  }
}

onMounted(async () => {
  try {
    const frame = await GetLatestFrameForROI()
    img.onload = async () => {
      imgWidth = img.naturalWidth || frame.width
      imgHeight = img.naturalHeight || frame.height
      imageLoaded.value = true
      loading.value = false
      await nextTick()
      resizeCanvas()
    }
    img.src = frame.image
  } catch (err) {
    loading.value = false
    console.error(err)
  }

  window.addEventListener('resize', resizeCanvas)
})

onUnmounted(() => {
  window.removeEventListener('resize', resizeCanvas)
})
</script>

<style scoped>
.roi-page {
  position: fixed;
  inset: 0;
  display: flex;
  flex-direction: column;
  background: radial-gradient(circle at top, #1e293b 0, #020617 55%);
  color: #e2e8f0;
  font-family: Inter, 'Segoe UI', system-ui, sans-serif;
}

.toolbar, .footer {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 22px;
  background: rgba(15, 23, 42, 0.92);
  border-color: #334155;
}

.toolbar { border-bottom: 1px solid #334155; }
.footer { border-top: 1px solid #334155; }

.title {
  font-size: 18px;
  font-weight: 900;
  color: #38bdf8;
}

.hint {
  margin-top: 4px;
  font-size: 12px;
  color: #94a3b8;
}

.zoom-badge, .roi-info {
  border: 1px solid #334155;
  background: #0f172a;
  border-radius: 999px;
  padding: 7px 14px;
  font-size: 12px;
  color: #93c5fd;
}

.stage {
  position: relative;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.canvas {
  width: 100%;
  height: 100%;
  cursor: crosshair;
  display: block;
}

.overlay-text {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  font-weight: 800;
}

.overlay-text.error { color: #f87171; }

.roi-info {
  display: flex;
  gap: 10px;
  border-radius: 8px;
}

.roi-info span { color: #64748b; }
.roi-info strong { color: #22c55e; }

.actions { display: flex; gap: 12px; }
.btn {
  border: none;
  border-radius: 8px;
  padding: 10px 24px;
  font-weight: 900;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn.secondary {
  background: transparent;
  color: #94a3b8;
  border: 1px solid #334155;
}

.btn.primary {
  background: linear-gradient(135deg, #2563eb, #06b6d4);
  color: white;
}

.btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
</style>
