<template>
  <div class="roi-page">
    <header class="toolbar">
      <div>
        <div class="title">{{ titleText }}</div>
        <div class="hint">{{ hintText }}</div>
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
      <div v-if="loading" class="overlay-text">正在识别最新相机帧...</div>
      <div v-else-if="!imageLoaded" class="overlay-text error">识别失败，检查棋盘格图片是否正确</div>
    </main>

    <footer class="footer">
      <div class="roi-info">
        <span>{{ footerLabel }}</span>
        <strong v-if="(isCornerCalibrationMode || isPoseCalibrationMode) && corners.length > 0">
          已检测 {{ corners.length }} 个角点
        </strong>
        <strong v-else-if="(isLineMode || isCalibrationMode) && line">
          {{ isCalibrationMode ? `长度: ${lineLength.toFixed(1)} 像素` : `(${Math.round(line.x1) }, ${Math.round(line.y1)}) → (${Math.round(line.x2) }, ${Math.round(line.y2)})` }}
        </strong>
        <strong v-else-if="!isLineMode && !isCalibrationMode && !isCornerCalibrationMode && !isPoseCalibrationMode && roi">
          X {{ Math.round(roi.x) }} / Y {{ Math.round(roi.y) }} / W {{ Math.round(roi.width) }} / H {{ Math.round(roi.height) }}
        </strong>
        <strong v-else>{{ emptyHintText }}</strong>
      </div>
      <div class="actions">
        <button class="btn secondary" @click="Window.Close()">×</button>
        <button class="btn primary" :disabled="(isCornerCalibrationMode || isPoseCalibrationMode) ? corners.length === 0 : ((isLineMode || isCalibrationMode) ? !line : !roi)" @click="confirm">√</button>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Window } from '@wailsio/runtime'
import * as HIKCameraService from '../../bindings/MINIMTSPRO/backend/hikcameraservice'
import { Events } from '@wailsio/runtime'

const route = useRoute()
const label = computed(() => route.query.label === 'B' ? 'B' : 'A')
const isLineMode = computed(() => route.query.mode === 'line')
const isCalibrationMode = computed(() => route.query.mode === 'calibration')
const isCornerCalibrationMode = computed(() => route.query.mode === 'calibration_corners')
const isPoseCalibrationMode = computed(() => route.query.mode === 'pose_calibration')
const calibrationRows = computed(() => {
  const value = parseInt(route.query.rows)
  return Number.isFinite(value) ? value : 7
})
const calibrationCols = computed(() => {
  const value = parseInt(route.query.cols)
  return Number.isFinite(value) ? value : 5
})
const calibrationSquareSize = computed(() => {
  const value = parseFloat(route.query.squareSize)
  return Number.isFinite(value) ? value : 25
})

const titleText = computed(() => {
  if (isPoseCalibrationMode.value) return '位姿标定 - 调整角点'
  if (isCornerCalibrationMode.value) return '相机标定 - 调整角点'
  if (isCalibrationMode.value) return '比例标定'
  if (isLineMode.value) return '绘制方向'
  return '选取标记 ' + label.value
})

const hintText = computed(() => {
  if (isPoseCalibrationMode.value || isCornerCalibrationMode.value) return '拖动角点调整位置，滚轮缩放图片'
  if (isCalibrationMode.value) return '左键拖拽绘制标定线段，滚轮缩放图片'
  if (isLineMode.value) return '左键拖拽绘制直线方向，滚轮缩放图片'
  return '左键拖拽绘制方框，滚轮缩放图片，确认后开始 DIC 追踪'
})

const footerLabel = computed(() => {
  if (isPoseCalibrationMode.value || isCornerCalibrationMode.value) return '棋盘格角点'
  if (isCalibrationMode.value) return '标定线段'
  if (isLineMode.value) return '直线'
  return 'ROI'
})

const emptyHintText = computed(() => {
  if (isPoseCalibrationMode.value || isCornerCalibrationMode.value) return '调整角点位置'
  if (isCalibrationMode.value) return '请绘制标定线段'
  if (isLineMode.value) return '请绘制直线'
  return '请绘制方框'
})

const stageRef = ref(null)
const canvasRef = ref(null)
const loading = ref(true)
const imageLoaded = ref(false)
const roi = ref(null)
const line = ref(null)
const corners = ref([])
const draggingCornerIndex = ref(-1)
const hoverCornerIndex = ref(-1)

const lineLength = computed(() => {
  if (!line.value) return 0
  const dx = line.value.x2 - line.value.x1
  const dy = line.value.y2 - line.value.y1
  return Math.sqrt(dx * dx + dy * dy)
})

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

const getCornerHitIndex = (point) => {
  const hitRadius = Math.max(10 / scale.value, 12)
  for (let i = 0; i < corners.value.length; i++) {
    const corner = corners.value[i]
    const dx = point.x - corner.x
    const dy = point.y - corner.y
    if (Math.sqrt(dx * dx + dy * dy) <= hitRadius) return i
  }
  return -1
}

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

  if (isCornerCalibrationMode.value || isPoseCalibrationMode.value) {
    corners.value.forEach((corner, index) => {
      const isDragging = draggingCornerIndex.value === index
      const isHovering = hoverCornerIndex.value === index
      const outerRadius = (isDragging ? 10 : isHovering ? 9 : 8) / scale.value
      const innerRadius = outerRadius * 0.55
      const strokeColor = isDragging ? '#f59e0b' : isHovering ? '#f97316' : '#ef4444'
      ctx.strokeStyle = strokeColor
      ctx.lineWidth = 3 / scale.value
      ctx.beginPath()
      ctx.stroke()
      ctx.beginPath()
      ctx.arc(corner.x, corner.y, innerRadius, 0, Math.PI * 2)
      ctx.stroke()

      ctx.lineWidth = 2 / scale.value
      ctx.beginPath()
      ctx.moveTo(corner.x - outerRadius * 0.8, corner.y)
      ctx.lineTo(corner.x + outerRadius * 0.8, corner.y)
      ctx.moveTo(corner.x, corner.y - outerRadius * 0.8)
      ctx.lineTo(corner.x, corner.y + outerRadius * 0.8)
      ctx.stroke()

      ctx.fillStyle = strokeColor
      ctx.font = `${12 / scale.value}px Inter, sans-serif`
      ctx.textAlign = 'center'
      ctx.fillText(String(index + 1), corner.x, corner.y - 12 / scale.value)
    })
  } else if (isLineMode.value || isCalibrationMode.value) {
    if (line.value) {
      ctx.strokeStyle = '#22c55e'
      ctx.lineWidth = 3 / scale.value
      ctx.beginPath()
      ctx.moveTo(line.value.x1, line.value.y1)
      ctx.lineTo(line.value.x2, line.value.y2)
      ctx.stroke()
      ctx.fillStyle = '#22c55e'
      ctx.beginPath()
      ctx.arc(line.value.x1, line.value.y1, 5 / scale.value, 0, Math.PI * 2)
      ctx.fill()
      ctx.beginPath()
      ctx.arc(line.value.x2, line.value.y2, 5 / scale.value, 0, Math.PI * 2)
      ctx.fill()
      drawArrow(ctx, line.value.x1, line.value.y1, line.value.x2, line.value.y2, '#22c55e')
    }
    if (drawing && startPoint && currentPoint) {
      ctx.strokeStyle = '#38bdf8'
      ctx.lineWidth = 3 / scale.value
      ctx.setLineDash([8 / scale.value, 4 / scale.value])
      ctx.beginPath()
      ctx.moveTo(startPoint.x, startPoint.y)
      ctx.lineTo(currentPoint.x, currentPoint.y)
      ctx.stroke()
      ctx.setLineDash([])
      ctx.fillStyle = '#38bdf8'
      ctx.beginPath()
      ctx.arc(startPoint.x, startPoint.y, 5 / scale.value, 0, Math.PI * 2)
      ctx.fill()
      ctx.beginPath()
      ctx.arc(currentPoint.x, currentPoint.y, 5 / scale.value, 0, Math.PI * 2)
      ctx.fill()
    }
  } else {
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
  }
  ctx.restore()
}

const drawArrow = (ctx, x1, y1, x2, y2, color) => {
  const dx = x2 - x1
  const dy = y2 - y1
  const len = Math.sqrt(dx * dx + dy * dy)
  if (len < 1) return
  const ux = dx / len
  const uy = dy / len
  const arrowLen = 20 / scale.value
  const arrowAngle = Math.PI / 6
  ctx.fillStyle = color
  ctx.beginPath()
  ctx.moveTo(x2, y2)
  ctx.lineTo(x2 - arrowLen * (ux * Math.cos(arrowAngle) - uy * Math.sin(arrowAngle)),
    y2 - arrowLen * (uy * Math.cos(arrowAngle) + ux * Math.sin(arrowAngle)))
  ctx.lineTo(x2 - arrowLen * (ux * Math.cos(arrowAngle) + uy * Math.sin(arrowAngle)),
    y2 - arrowLen * (uy * Math.cos(arrowAngle) - ux * Math.sin(arrowAngle)))
  ctx.closePath()
  ctx.fill()
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

  if (isCornerCalibrationMode.value || isPoseCalibrationMode.value) {
    const point = canvasPointToImagePoint(event.clientX, event.clientY)
    const hitIndex = getCornerHitIndex(point)
    if (hitIndex >= 0) {
      draggingCornerIndex.value = hitIndex
      hoverCornerIndex.value = hitIndex
      drawing = true
      currentPoint = point
      canvasRef.value.style.cursor = 'grabbing'
      return
    }
    return
  }

  drawing = true
  startPoint = clampPoint(canvasPointToImagePoint(event.clientX, event.clientY))
  currentPoint = startPoint
  draw()
}

const handleMouseMove = (event) => {
  if (!drawing) return

  if ((isCornerCalibrationMode.value || isPoseCalibrationMode.value) && draggingCornerIndex.value >= 0) {
    const point = clampPoint(canvasPointToImagePoint(event.clientX, event.clientY))
    corners.value[draggingCornerIndex.value] = point
    currentPoint = point
    hoverCornerIndex.value = draggingCornerIndex.value
    canvasRef.value.style.cursor = 'grabbing'
    draw()
    return
  }

  if (isCornerCalibrationMode.value || isPoseCalibrationMode.value) {
    const point = canvasPointToImagePoint(event.clientX, event.clientY)
    hoverCornerIndex.value = getCornerHitIndex(point)
    canvasRef.value.style.cursor = hoverCornerIndex.value >= 0 ? 'grab' : 'default'
    draw()
    return
  }

  currentPoint = clampPoint(canvasPointToImagePoint(event.clientX, event.clientY))
  draw()
}

const handleMouseUp = () => {
  if (isCornerCalibrationMode.value || isPoseCalibrationMode.value) {
    draggingCornerIndex.value = -1
    drawing = false
    canvasRef.value.style.cursor = hoverCornerIndex.value >= 0 ? 'grab' : 'default'
    return
  }

  if (!drawing || !startPoint || !currentPoint) return
  drawing = false

  if (isLineMode.value || isCalibrationMode.value) {
    const dx = currentPoint.x - startPoint.x
    const dy = currentPoint.y - startPoint.y
    const len = Math.sqrt(dx * dx + dy * dy)
    if (len >= 5) {
      line.value = {
        x1: startPoint.x,
        y1: startPoint.y,
        x2: currentPoint.x,
        y2: currentPoint.y
      }
    }
  } else {
    const x = Math.min(startPoint.x, currentPoint.x)
    const y = Math.min(startPoint.y, currentPoint.y)
    const width = Math.abs(currentPoint.x - startPoint.x)
    const height = Math.abs(currentPoint.y - startPoint.y)
    if (width >= 8 && height >= 8) {
      roi.value = { x, y, width, height }
    }
  }
  startPoint = null
  currentPoint = null
  draw()
}

const confirm = async () => {
  if (isPoseCalibrationMode.value) {
    if (corners.value.length === 0) return
    try {
      console.log("corners:", corners.value)
      await HIKCameraService.AddPoseCalibration(
        corners.value,
        calibrationRows.value,
        calibrationCols.value,
        calibrationSquareSize.value
      )
      await Window.Close()
    } catch (err) {
      alert(`位姿标定失败：${err}`)
    }
    return
  }


  if (isCornerCalibrationMode.value) {
    if (corners.value.length === 0) return
    try {
      const result = await HIKCameraService.AddCalibrationCorners(
        corners.value,
        calibrationRows.value,
        calibrationCols.value,
        calibrationSquareSize.value
      )
      if (result.success) {
        Events.Emit('calibration_added', { data: { count: result.count } })
        await Window.Close()
      } else {
        alert(`标定失败：${result.error}`)
      }
    } catch (err) {
      alert(`标定失败：${err}`)
    }
    return
  }
  
  if (isCalibrationMode.value) {
    if (!line.value) return
    Events.Emit('hik_calibration_selected', { length: lineLength.value })
    await Window.Close()
    return
  }

  if (isLineMode.value || isCalibrationMode.value) {
    if (!line.value) return
    try {
      await HIKCameraService.SetDirectionLine(line.value)
      await Window.Close()
    } catch (err) {
      alert(`方向设置失败：${err}`)
    }
  } else {
    if (!roi.value) return
    try {
      await HIKCameraService.SetROI(label.value, roi.value)
      await Window.Close()
    } catch (err) {
      alert(`ROI 设置失败：${err}`)
    }
  }
  
}

onMounted(async () => {
  try {
    if (isCornerCalibrationMode.value || isPoseCalibrationMode.value) {
      const cornersStr = route.query.corners
      const imageStr = route.query.image
      if (cornersStr && imageStr) {
        corners.value = JSON.parse(decodeURIComponent(cornersStr))
        img.onload = async () => {
          imgWidth = img.naturalWidth
          imgHeight = img.naturalHeight
          imageLoaded.value = true
          loading.value = false
          await nextTick()
          resizeCanvas()
        }
        img.src = decodeURIComponent(imageStr)
      } else {
        const result = await HIKCameraService.FindChessboardCorners(calibrationRows.value, calibrationCols.value)
        if (!result?.success) {
          loading.value = false
          return
        }
        corners.value = result.corners || []
        img.onload = async () => {
          imgWidth = img.naturalWidth
          imgHeight = img.naturalHeight
          imageLoaded.value = true
          loading.value = false
          await nextTick()
          resizeCanvas()
        }
        img.src = result.image
      }
    } else {
      const frame = await HIKCameraService.GetLatestFrameForROI()
      img.onload = async () => {
        imgWidth = img.naturalWidth || frame.width
        imgHeight = img.naturalHeight || frame.height
        imageLoaded.value = true
        loading.value = false
        await nextTick()
        resizeCanvas()
      }
      img.src = frame.image
    }
  } catch (err) {
    loading.value = false
    console.error(err)
  }

  window.addEventListener('resize', resizeCanvas)
  if (canvasRef.value) canvasRef.value.style.cursor = 'default'
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
