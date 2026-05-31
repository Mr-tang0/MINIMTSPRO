<template>
  <div 
    class="metric-card"
    :style="{ '--accent': accent }"
    @dblclick="handleDoubleClick"
  >
    <!-- 左侧胶囊型霓虹指示条 -->
    <div class="card-accent-bar"></div>
    
    <div class="card-content">
      <!-- 水平居中对称排列：mainLabel（左）、mainValue（中）、mainUnit（右） -->
      <div class="main-row">
        <span class="main-label">{{ mainLabel }}</span>
        <span class="main-value">{{ mainValue.toFixed(mainPrecision) }}</span>
        <span class="main-unit">{{ mainUnit }}</span>
      </div>
      
      <!-- 小字副数据区（可选） -->
      <div v-if="showSub && subValue !== undefined" class="sub-row">
        <span class="sub-label">{{ subLabel }}</span>
        <span class="sub-value">{{ typeof subValue === 'string' ? subValue : subValue.toFixed(subPrecision) }}</span>
        <span class="sub-unit">{{ subUnit }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  mainLabel: { type: String, default: '' },
  mainValue: { type: Number, default: 0 },
  mainUnit: { type: String, default: '' },
  mainPrecision: { type: Number, default: 2 },
  subLabel: { type: String, default: '' },
  subValue: { type: Number || String, default: undefined },
  subUnit: { type: String, default: '' },
  subPrecision: { type: Number, default: 2 },
  showSub: { type: Boolean, default: true },
  accent: { type: String, default: '#3b82f6' }
});

const emit = defineEmits(['dblclick']);

const handleDoubleClick = () => {
  emit('dblclick', props.mainLabel);
};
</script>

<style scoped>
.metric-card {
  background: #1e293b;
  border-radius: 10px;
  padding: 14px 16px;
  position: relative;
  overflow: hidden;
  border: 1px solid #334155;
  cursor: pointer;
  transition: all 0.2s ease;
}

.metric-card:hover {
  border-color: var(--accent);
}

.metric-card:active {
  transform: scale(0.98);
}

/* 左侧指示条 */
.card-accent-bar {
  position: absolute;
  top: 50%;
  left: 8px;
  transform: translateY(-50%);
  width: 4px;
  height: 80%;
  min-height: 30px;
  /* max-height: 42px; */
  background: var(--accent);
  border-radius: 2px;
}

.card-content {
  position: relative;
  z-index: 1;
}

/* 主数据行 - 水平居中对称排列 */
.main-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
}

.main-label {
  font-size: 18px;
  font-weight: 600;
  color: #94a3b8;
  flex-shrink: 0;
  margin-left: 14px;
}

.main-value {
  font-size: 32px;
  font-weight: 900;
  color: #f8fafc;
  font-family: 'Monaco', 'Consolas', monospace;
  flex: 1;
  text-align: center;
  min-width: 0;
}

.main-unit {
  font-size: 18px;
  color: #64748b;
  flex-shrink: 0;
}

/* 副数据行 */
.sub-row {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 6px;
  padding-top: 8px;
  border-top: 1px solid rgba(71, 85, 105, 0.3);
}

.sub-label {
  font-size: 16px;
  color: #64748b;
}

.sub-value {
  font-size: 16px;
  color: #94a3b8;
  font-family: 'Monaco', 'Consolas', monospace;
}

.sub-unit {
  font-size: 16px;
  color: #64748b;
}
</style>