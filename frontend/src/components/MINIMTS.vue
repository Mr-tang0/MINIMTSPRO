<template>
  <div class="dashboard">
    <!-- 侧边导航栏 -->
    <aside class="sidebar">
      <div class="sidebar-top">
        <div class="sidebar-header">
          <div class="logo-icon">
            <img src="../res/user.png" alt="Logo" />
          </div>
          <div class="badge">1</div>
          
          <span class="logo-text">MTS</span>
        </div>

        <nav class="nav-icons">
          <div 
            class="icon-wrapper" 
            :class="{ active: ModalStatus.showConnectModal }" 
            @click="ModalStatus.showConnectModal = true"
          >
            <i class="ri-link-m icon"></i>
            <span class="btn-tip">连接</span>
          </div>
          
          <div 
            class="icon-wrapper" 
            :class="{ active: ModalStatus.showProjectWindow }" 
            @click="ModalStatus.showProjectWindow = true"
          >
            <i class="ri-file-list-3-line icon"></i>
            <span class="btn-tip">项目</span>
          </div>
          
          <div class="icon-wrapper">
            <i class="ri-folder-2-line icon"></i>
            <span class="btn-tip">文件</span>
          </div>
          
          <div v-if="SystemStatus.CameraOpened" class="icon-wrapper" @click="CallHIKCameraWindow">
            <i class="ri-camera-lens-line icon"></i>
            <span class="btn-tip">相机</span>
          </div>
        </nav>
      </div>

      <div class="sidebar-footer">
        <div class="icon-wrapper">
          <i class="ri-settings-4-line icon"></i>
          <span class="btn-tip">设置</span>
        </div>
      </div>
    </aside>

    <!-- 主内容区域 -->
    <main class="main-area">
      <!-- 顶部状态栏 -->
      <header class="top-bar">
        <div class="greeting">
          <h2 class="app-title">MINIMTS <span class="version-badge">PRO</span></h2>
          <!-- <p class="subtitle">材料试验机控制系统 · 精密测试环境</p> -->
        </div>
        <div class="global-status">
          <div class="connection-badge" :class="{ connected: SystemStatus.MINIMTSOpened}">
            <span class="status-dot"></span>
            <span>MINIMTS {{ SystemStatus.MINIMTSOpened?'已连接':'未连接' }}</span>
          </div>
          <div class="connection-badge camera" :class="{ connected: DataValues.cameraConnected}">
            <span class="status-dot"></span>
            <span>相机 {{ DataValues.cameraConnected?'已连接':'未连接' }}</span>
          </div>
        </div>
      </header>

      <!-- 数据卡片网格 -->
      <section class="metrics-grid">
        <!-- 应力卡片 -->
        <MetricCard
          main-label="应力"
          :main-value="DataValues.stress"
          main-unit="MPa"
          :main-precision="1"
          sub-label="载荷"
          :sub-value="DataValues.load"
          sub-unit="N"
          :sub-precision="1"
          accent="#f59e0b"
          @dblclick="handleCardDoubleClick('load')"
        />
        
        <!-- 应变卡片 -->
        <MetricCard
          main-label="应变"
          :main-value="DataValues.strain"
          main-unit="ε"
          :main-precision="3"
          sub-label="位移"
          :sub-value="DataValues.disp"
          sub-unit="mm"
          :sub-precision="3"
          accent="#8b5cf6"
          @dblclick="handleCardDoubleClick('disp')"
        />
        
        <!-- 视频应变卡片 -->
        <MetricCard
          main-label="视频应变"
          :main-value="DataValues.videoStrain"
          main-unit="ε"
          :main-precision="3"
          sub-label="视频位移"
          :sub-value="DataValues.videoDisp"
          sub-unit="mm"
          :sub-precision="3"
          accent="#06b6d4"
          @dblclick="handleCardDoubleClick('videoDisp')"
        />
        
        <!-- 时间卡片 -->
        <MetricCard
          main-label="时间"
          :main-value="DataValues.time"
          main-unit="S"
          :main-precision="1"
          sub-label="北京时间"
          :sub-precision="0"
          sub-unit=""
          :sub-value="beijingTime"
          accent="#ec4899"
          @dblclick="handleCardDoubleClick('time')"
        />
      </section>

      <!-- 图表与控制混合区域 -->
      <div class="bottom-panel">
        <section class="chart-section">
          <div class="chart-header">
            <h3 class="section-title">
              <i class="ri-bar-chart-2-line"></i> 实时数据曲线
            </h3>
            <div class="view-tabs">
              <!-- 按钮1：载荷-时间曲线 -->
              <button 
                class="tab-btn" 
                :class="{ active: currentView === 'load_time' }"
                @click="currentView = 'load_time'; refreshChartUI()"
              >
                载荷-时间
              </button>
              <!-- 按钮2：位移-时间与视频位移-时间 -->
              <button 
                class="tab-btn" 
                :class="{ active: currentView === 'disp_video' }"
                @click="currentView = 'disp_video'; refreshChartUI()"
              >
                位移-时间
              </button>
              <!-- 按钮3：应变-时间与视频应变 -->
              <button 
                class="tab-btn" 
                :class="{ active: currentView === 'strain_video' }"
                @click="currentView = 'strain_video'; refreshChartUI()"
              >
                应变-时间
              </button>
              <!-- 按钮4：应力-应变与应力-视频应变 -->
              <button 
                class="tab-btn" 
                :class="{ active: currentView === 'stress_strain' }"
                @click="currentView = 'stress_strain'; refreshChartUI()"
              >
                应力-应变
              </button>
            </div>
          </div>
          <div ref="chartRef" class="chart-container"></div>
        </section>

        <!-- 右侧紧凑控制区 -->
        <aside class="quick-controls">
          <div class="control-group">
            <h4 class="group-title"><i class="ri-dashboard-3-line"></i> 试验控制</h4>
            <div class="action-buttons">
              <button class="action-btn" :class="{'start': !isTesting, 'stop': isTesting}" @click="toggleTest">
                <i class="ri-play-fill"></i> 
                {{ isTesting ? '停止' : '开始' }}
              </button>
              <!-- <button class="action-btn stop" @click="stopTest" :disabled="!isTesting">
                <i class="ri-stop-fill"></i> 停止
              </button> -->
            </div>
          </div>

          <div class="control-group">
            <h4 class="group-title"><i class="ri-database-line"></i> 数据操作</h4>
            <div class="data-buttons">
              <button class="data-btn" @click="ClearCharts">
                <i class="ri-trash-2-line"></i> 清除画面
              </button>
              <button class="data-btn" @click="saveData">
                <i class="ri-save-line"></i> 保存数据
              </button>
              <button class="data-btn" @click="resetDisp">
                <i class="ri-arrow-left-right-line"></i> 位移归零
              </button>
            </div>
          </div>

          <div class="control-group">
            <h4 class="group-title"><i class="ri-equalizer-line"></i> 手动操作</h4>
            <div class="speed-input-row">
              <label>速度 (mm/s)</label>
              <input type="number" step="0.1" v-model="SystemStatus.JogSpeed" class="speed-input" />
            </div>
            <div class="jog-buttons">
              <button class="jog-btn up"  @mousedown="jog(SystemStatus.JogSpeed)" @mouseup="jog(0)">
                <i class="ri-arrow-up-s-line"></i> 拉向
              </button>
              <button class="jog-btn halt" @click="jog(0)">
                <i class="ri-stop-fill"></i> 停止
              </button>
              <button class="jog-btn down"  @mousedown="jog(-SystemStatus.JogSpeed)" @mouseup="jog(0)">
                <i class="ri-arrow-down-s-line"></i> 压向
              </button>
            </div>
          </div>

          

        </aside>
      </div>
    </main>

    <!-- 连接模态框 -->
    <Teleport to="body">
      <div v-if="ModalStatus.showConnectModal" class="modal-overlay" @click.self="ModalStatus.showConnectModal = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3><i class="ri-link-m"></i> 设备连接</h3>
            <button class="close-btn" @click="ModalStatus.showConnectModal = false">
              <i class="ri-close-line"></i>
            </button>
          </div>
          <div class="modal-body">
            <div class="device-panel">
              <div class="panel-title">
                <span><i class="ri-serial-port-line"></i> MINIMTS 控制器</span>
                <button class="refresh-btn" @click="refreshMINIMTSDevices" :disabled="SystemStatus.MINIMTSRefreshing">
                  <i class="ri-refresh-line" :class="{ spinning: SystemStatus.MINIMTSRefreshing }"></i>
                </button>
              </div>
              <select v-model="SystemStatus.SelectedMINIMTS" class="device-select">
                <option value="">请选择串口</option>
                <option v-for="d in MINIMTSDevices" :key="d">{{ d }}</option>
              </select>
              <button 
                class="connect-btn" 
                :class="{ connected: SystemStatus.MINIMTSOpened }"
                @click="handleMINIMTSConnect"
                :disabled="!SystemStatus.SelectedMINIMTS"
              >
                {{ SystemStatus.MINIMTSOpened ? '断开设备' : '连接设备' }}
              </button>
            </div>
            
            <div class="device-panel">
              <div class="panel-title">
                <span><i class="ri-camera-line"></i> 视频引伸计</span>
                <button class="refresh-btn" @click="refreshCameraDevices" :disabled="SystemStatus.CameraRefreshing">
                  <i class="ri-refresh-line" :class="{ spinning: SystemStatus.CameraRefreshing }"></i>
                </button>
              </div>
              <select v-model="SystemStatus.SelectedCamera" class="device-select">
                <option value="">请选择相机</option>
                <option v-for="d in CameraDevices" :key="d">{{ d }}</option>
              </select>
              <button 
                class="connect-btn" 
                :class="{ connected: SystemStatus.CameraOpened }"
                @click="handleCameraConnect"
                :disabled="!SystemStatus.SelectedCamera"
              >
                {{ SystemStatus.CameraOpened ? '断开相机' : '连接相机' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 项目设置模态框 -->
    <Teleport to="body">
      <div v-if="ModalStatus.showProjectWindow" class="modal-overlay">
        <div class="modal-container project-modal">
          <div class="modal-header">
            <h2>新建试验项目</h2>
            <button class="close-btn" @click="ModalStatus.showProjectWindow = false">
              <i class="ri-close-line"></i>
            </button>
          </div>
          <div class="modal-body">
            <Project 
              :isCameraConnected="SystemStatus.CameraOpened" 
              @cancel="ModalStatus.showProjectWindow = false"
              @submit="handleProjectSubmit"
            />
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue';
import { Events } from '@wailsio/runtime';
import * as echarts from 'echarts';
import MetricCard from './MetricCard.vue';
import Project from './Project.vue';
import { 
  CallHIKCameraWindow,
  CloseHIKCameraWindow,
  GetMINIMTSDevices,
  OpenMINIMTS,
  CloseMINIMTS,
  JogMove,
  ClearDataCache,
  StartMeasurement,
  StopMeasurement,
  CallDataToZero,
  SaveDataTCsv,
} from "../../bindings/changeme/backend/minimtsservice";

import { 
  GetHIKCameraDevices,
  OpenHIKCamera,
  CloseHIKCamera,
} from "../../bindings/changeme/backend/hikcameraservice";


// 系统状态
const SystemStatus = reactive({
  MINIMTSRefreshing: false,
  CameraRefreshing: false,
  SelectedMINIMTS: '',
  SelectedCamera: '',
  MINIMTSOpened: false,
  CameraOpened: false,
  JogSpeed: 0.1
});

const MINIMTSDevices = ref([]);
const CameraDevices = ref([]);

// 模态框状态
const ModalStatus = reactive({
  showConnectModal: true, //打开软件时候，显示连接窗口
  showProjectWindow: false,
  showSystemWindow: false,
});

// 数据值
const DataValues = reactive({
  load: 0.0,
  stress: 0.0,
  disp: 0.0,
  videoDisp: 0.0,
  strain: 0.0,
  videoStrain: 0.0,
  time: 0.0,
  cameraConnected: false,
  minimtsConnected: false
});

// 北京时间
const beijingTime = ref('');

// 更新北京时间
const updateBeijingTime = () => {
  const now = new Date();
  const hours = String(now.getHours()).padStart(2, '0');
  const minutes = String(now.getMinutes()).padStart(2, '0');
  const seconds = String(now.getSeconds()).padStart(2, '0');
  beijingTime.value = `${hours}:${minutes}:${seconds}`;
};

// 曲线数据序列
const DataSerials = reactive({
  load: [],
  stress: [],
  disp: [],
  videoDisp: [],
  strain: [],
  videoStrain: [],
  time: []
});

// 图表相关
const chartRef = ref(null);
let myChart = null;
const currentView = ref('load_time');
const isTesting = ref(false);



// 刷新设备列表
const refreshMINIMTSDevices = async() => {
  try {
    SystemStatus.MINIMTSRefreshing = true;
    const devices = await GetMINIMTSDevices();
    MINIMTSDevices.value = devices;
    if(!SystemStatus.MINIMTSOpened && devices.length > 0) {
      SystemStatus.SelectedMINIMTS = devices[devices.length - 1];
    }
  } catch (error) {
    alert(error.message);
  } finally {
    SystemStatus.MINIMTSRefreshing = false;
  }
};

const refreshCameraDevices = async () => {
  try { 
    const devices = await GetHIKCameraDevices();
    CameraDevices.value = devices;
    if(!SystemStatus.CameraOpened && devices.length > 0) {
      SystemStatus.SelectedCamera = devices[devices.length - 1];
    }
  } catch (error) {
    alert(error.message);
  } finally {
    SystemStatus.CameraRefreshing = false;
  }
  
};

// 设备连接
const handleMINIMTSConnect = async () => { 
  if (SystemStatus.MINIMTSOpened) {
    try { 
      await CloseMINIMTS();
      SystemStatus.MINIMTSOpened = false;
    } catch (error) {
      alert(error.message);
    }
  } else {
    try { 
      await OpenMINIMTS(SystemStatus.SelectedMINIMTS);
      SystemStatus.MINIMTSOpened = true;
    } catch (error) {
      alert(error.message);
    }
  }
};

const handleCameraConnect = async () => { 
  if (SystemStatus.CameraOpened) {
    try { 
      await CloseHIKCamera();
      await CloseHIKCameraWindow();
      SystemStatus.CameraOpened = false;
    } catch (error) {
      alert(error.message);
    }
  } else {
    try { 
      await OpenHIKCamera(SystemStatus.SelectedCamera);
      await CallHIKCameraWindow();
      SystemStatus.CameraOpened = true;
    } catch (error) {
      alert(error.message);
    }
  }
};

const jog = async(speed) => { 
  try { 
    await JogMove(speed);
  } catch (error) {
    alert(error.message);
  }
};

// 卡片双击事件
const handleCardDoubleClick = async (key) => {
  try {
    await CallDataToZero(key);
  } catch (error) {
    alert(error.message);
  }
};

// 图表初始化
const initChart = () => {
  if (!chartRef.value) return;
  if (myChart) myChart.dispose();
  
  myChart = echarts.init(chartRef.value);
  myChart.setOption({
    backgroundColor: '#ffffff',
    animation: false,
    grid: { top: 30, left: 50, right: 50, bottom: 35, containLabel: true },
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(30, 41, 59, 0.95)',
      borderColor: '#475569',
      textStyle: { color: '#f8fafc' },
      axisPointer: { type: 'cross', lineStyle: { color: '#94a3b8' } }
    },
    legend: {
      data: [],
      textStyle: { color: '#374151' },
      top: 5
    },
    xAxis: {
      type: 'value',
      scale: true,
      axisLine: { lineStyle: { color: '#d1d5db' } },
      axisTick: { show: true, lineStyle: { color: '#d1d5db' } },
      splitLine: { lineStyle: { color: '#e5e7eb', type: 'solid', opacity: 1 } },
      axisLabel: { 
        color: '#6b7280', 
        fontSize: 12,
        formatter: (val) => val.toFixed(2)
      }
    },
    yAxis: {
      type: 'value',
      scale: true,
      axisLine: { show: true, lineStyle: { color: '#d1d5db' } },
      axisTick: { show: true, lineStyle: { color: '#d1d5db' } },
      splitLine: { lineStyle: { color: '#e5e7eb', type: 'solid', opacity: 1 } },
      axisLabel: { 
        color: '#6b7280', 
        fontSize: 12,
        formatter: (val) => val.toFixed(3)
      }
    },
    series: []
  });
};

// 刷新图表
const refreshChartUI = () => {
  if (!myChart) return;
  
  const series = [];
  const legendData = [];
  let xMin = 0, xMax = 0, yMin = 0, yMax = 0;
  
  switch(currentView.value) {
    case 'load_time':
      legendData.push('载荷');
      series.push({
          name: '载荷',
          type: 'line',
          smooth: true,
          symbol: 'none',
          data: DataSerials.time.map((t, i) => [t, DataSerials.load[i] || 0]),
          lineStyle: { color: '#3b82f6', width: 2 },
          itemStyle: { color: '#3b82f6' }
        });
      if (DataSerials.time.length > 0) {
        xMin = Math.min(...DataSerials.time);
        xMax = Math.max(...DataSerials.time);
        yMin = Math.min(...DataSerials.load);
        yMax = Math.max(...DataSerials.load);
      }
      break;
      
    case 'disp_video':
      legendData.push('位移');
      series.push({
          name: '位移',
          type: 'line',
          smooth: true,
          symbol: 'none',
          data: DataSerials.time.map((t, i) => [t, DataSerials.disp[i] || 0]),
          lineStyle: { color: '#10b981', width: 2 },
          itemStyle: { color: '#10b981' }
        });
      if (SystemStatus.CameraOpened) {
        legendData.push('视频位移');
        series.push({
          name: '视频位移',
          type: 'line',
          smooth: true,
          symbol: 'none',
          data: DataSerials.time.map((t, i) => [t, DataSerials.videoDisp[i] || 0]),
          lineStyle: { color: '#3b82f6', width: 2 },
          itemStyle: { color: '#3b82f6' }
        });
      }
      if (DataSerials.time.length > 0) {
        xMin = Math.min(...DataSerials.time);
        xMax = Math.max(...DataSerials.time);
        const allDisp = SystemStatus.CameraOpened 
          ? [...DataSerials.disp, ...DataSerials.videoDisp]
          : DataSerials.disp;
        yMin = Math.min(...allDisp);
        yMax = Math.max(...allDisp);
      }
      break;
      
    case 'strain_video':
      legendData.push('应变');
      series.push({
          name: '应变',
          type: 'line',
          smooth: true,
          symbol: 'none',
          data: DataSerials.time.map((t, i) => [t, DataSerials.strain[i] || 0]),
          lineStyle: { color: '#8b5cf6', width: 2 },
          itemStyle: { color: '#8b5cf6' }
        });
      if (SystemStatus.CameraOpened) {
        legendData.push('视频应变');
        series.push({
          name: '视频应变',
          type: 'line',
          smooth: true,
          symbol: 'none',
          data: DataSerials.time.map((t, i) => [t, DataSerials.videoStrain[i] || 0]),
          lineStyle: { color: '#ec4899', width: 2 },
          itemStyle: { color: '#ec4899' }
        });
      }
      if (DataSerials.time.length > 0) {
        xMin = Math.min(...DataSerials.time);
        xMax = Math.max(...DataSerials.time);
        const allStrain = SystemStatus.CameraOpened
          ? [...DataSerials.strain, ...DataSerials.videoStrain]
          : DataSerials.strain;
        yMin = Math.min(...allStrain);
        yMax = Math.max(...allStrain);
      }
      break;
      
    case 'stress_strain':
      legendData.push('应力-应变');
      series.push({
          name: '应力-应变',
          type: 'line',
          smooth: true,
          symbol: 'none',
          data: DataSerials.strain.map((s, i) => [s, DataSerials.stress[i] || 0]),
          lineStyle: { color: '#f59e0b', width: 2 },
          itemStyle: { color: '#f59e0b' }
        });
      if (SystemStatus.CameraOpened) {
        legendData.push('应力-视频应变');
        series.push({
          name: '应力-视频应变',
          type: 'line',
          smooth: true,
          symbol: 'none',
          data: DataSerials.videoStrain.map((s, i) => [s, DataSerials.stress[i] || 0]),
          lineStyle: { color: '#06b6d4', width: 2 },
          itemStyle: { color: '#06b6d4' }
        });
      }
      if (DataSerials.strain.length > 0 || (SystemStatus.CameraOpened && DataSerials.videoStrain.length > 0)) {
        const allStrainValues = SystemStatus.CameraOpened
          ? [...DataSerials.strain, ...DataSerials.videoStrain].filter(v => v !== undefined)
          : DataSerials.strain.filter(v => v !== undefined);
        const allStressValues = DataSerials.stress.filter(v => v !== undefined);
        if (allStrainValues.length > 0) {
          xMin = Math.min(...allStrainValues);
          xMax = Math.max(...allStrainValues);
        }
        if (allStressValues.length > 0) {
          yMin = Math.min(...allStressValues);
          yMax = Math.max(...allStressValues);
        }
      }
      break;
  }
  
  myChart.setOption({
    legend: { data: legendData },
    series: series,
    xAxis: {
      min: xMin,
      max: xMax
    },
    yAxis: {
      min: yMin,
      max: yMax
    }
  });
};

// 清除图表数据
const ClearCharts = async () => {
  await ClearDataCache();
  Object.keys(DataSerials).forEach(k => DataSerials[k] = []);
  refreshChartUI();
};

// 处理项目提交
const handleProjectSubmit = (form) => {
  console.log('Project submitted:', form);
  ModalStatus.showProjectWindow = false;
  // alert('项目配置已保存');
};

// 保存数据
const saveData = async () => {
  try {
    await SaveDataTCsv();
  } catch (error) {
    console.error('保存数据失败:', error);
  }
};

// 位移归零
const resetDisp = () => {
  console.log('Resetting displacement...');
  alert('位移归零功能开发中');
};

// 测试控制
const toggleTest = async () => { 
  if (isTesting.value) {
    try {
      await StopMeasurement();
      isTesting.value = false;
    } catch (error) {
      console.error('停止测量失败:', error);
    }

  }else {
    try {
      await StartMeasurement();
      isTesting.value = true;
    } catch (error) {
      console.error('启动测量失败:', error);
    }
  }
};
const emergencyStop = () => { 
  isTesting.value = false; 
  jog(0);
};

onMounted(async () => {
  initChart();
  refreshMINIMTSDevices();
  refreshCameraDevices();
  refreshChartUI();
  
  // 初始化北京时间
  updateBeijingTime();
  // 每秒更新北京时间
  const timeInterval = setInterval(updateBeijingTime, 1000);
  
  window.addEventListener('resize', () => myChart?.resize());

  Events.On('update_status', (status) => {
    const data = status.data;
    
    if (data.load !== undefined) DataValues.load = data.load;
    if (data.stress !== undefined) DataValues.stress = data.stress;
    if (data.disp !== undefined) DataValues.disp = data.disp;
    if (data.videoDisp !== undefined) DataValues.videoDisp = data.videoDisp;
    if (data.strain !== undefined) DataValues.strain = data.strain;
    if (data.videoStrain !== undefined) DataValues.videoStrain = data.videoStrain;
    if (data.time !== undefined) DataValues.time = data.time;
    if (data.cameraConnected !== undefined) DataValues.cameraConnected = data.cameraConnected;
    if (data.minimtsConnected !== undefined) DataValues.minimtsConnected = data.minimtsConnected;
    
    DataSerials.time.push(data.time || DataSerials.time.length * 0.05);
    DataSerials.load.push(data.load || 0);
    DataSerials.stress.push(data.stress || 0);
    DataSerials.disp.push(data.disp || 0);
    DataSerials.videoDisp.push(data.videoDisp || 0);
    DataSerials.strain.push(data.strain || 0);
    DataSerials.videoStrain.push(data.videoStrain || 0);
    
    refreshChartUI();
  });

});

onUnmounted(() => {
  if (myChart) myChart.dispose();
  clearInterval(timeInterval);
});
</script>

<style scoped>
@import url('https://cdn.jsdelivr.net/npm/remixicon@3.5.0/fonts/remixicon.css');

.dashboard {
  --bg-main: #1e293b;
  --bg-sidebar: #0f172a;
  --bg-panel: #273549;
  --bg-card: #334155;
  --border-color: #475569;
  --text-primary: #f8fafc;
  --text-regular: #cbd5e1;
  --text-muted: #94a3b8;
  --accent-blue: #3b82f6;
  --accent-hover: #2563eb;
  
  --danger: #ef4444;
  --danger-hover: #d32f2f;
  --success: #10b981;

  display: flex;
  position: fixed;
  inset: 0;
  background: var(--bg-main);
  color: var(--text-regular);
  font-family: 'Inter', 'Segoe UI', system-ui, sans-serif;
  overflow: hidden;
}

.sidebar {
  width: 90px;
  background: var(--bg-sidebar);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 0;
  border-right: 1px solid var(--border-color);
  justify-content: space-between;
  box-shadow: 2px 0 10px rgba(0, 0, 0, 0.1);
  z-index: 10;
}

.sidebar-top {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}

.sidebar-header {
  position: relative;
  width: 52px;
  height: 52px;
  cursor: pointer;
  margin-bottom: 30px;
}

.logo-icon {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 10px rgba(59,130,246,0.3);
  overflow: hidden;
}

.logo-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.badge {
  position: absolute;
  right: -4px;
  bottom: -4px;
  min-width: 20px;
  height: 20px;
  padding: 0 5px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: bold;
  background: var(--danger);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--bg-sidebar);
  z-index: 10;
}

.logo-text {
  display: block;
}

.nav-icons {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  position: relative;
  transition: all 0.2s;
  background: transparent;
}

.icon-wrapper:hover {
  background: rgba(255,255,255,0.05);
}

.icon-wrapper.active {
  background: var(--accent-blue);
}

.icon-wrapper.active .icon {
  color: white;
}

.icon {
  font-size: 20px;
  color: var(--text-muted);
}

.btn-tip {
  position: absolute;
  left: calc(100% + 8px);
  padding: 4px 8px;
  background: var(--bg-card);
  border-radius: 4px;
  font-size: 12px;
  white-space: nowrap;
  opacity: 0;
  visibility: hidden;
  transition: all 0.2s;
  color: var(--text-primary);
  z-index: 100;
}

.icon-wrapper:hover .btn-tip {
  opacity: 1;
  visibility: visible;
}

.sidebar-footer {
  padding-bottom: 20px;
}

.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: var(--bg-sidebar);
  border-bottom: 1px solid var(--border-color);
}

.greeting .app-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.version-badge {
  font-size: 10px;
  padding: 2px 6px;
  background: var(--accent-blue);
  border-radius: 4px;
  margin-left: 8px;
}

.subtitle {
  font-size: 13px;
  color: var(--text-muted);
  margin: 4px 0 0 0;
}

.connection-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 20px;
  color: var(--danger);
  font-size: 13px;
}

.connection-badge.connected {
  background: rgba(16, 185, 129, 0.1);
  color: var(--success);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--danger);
}

.connection-badge.connected .status-dot {
  background: var(--success);
  box-shadow: 0 0 8px var(--success);
}

.connection-badge.camera.connected {
  background: rgba(6, 182, 212, 0.1);
  color: #06b6d4;
}

.connection-badge.camera.connected .status-dot {
  background: #06b6d4;
  box-shadow: 0 0 8px #06b6d4;
}

/* 数据卡片网格 */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  padding: 20px 24px;
}

/* 图表区域 */
.bottom-panel {
  flex: 1;
  display: flex;
  gap: 20px;
  padding: 0 24px 24px;
  min-height: 300px;
}

.chart-section {
  flex: 1;
  background: var(--bg-panel);
  border-radius: 12px;
  padding: 16px;
  border: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.view-tabs {
  display: flex;
  gap: 8px;
}

.tab-btn {
  padding: 1px 10px;
  min-width: 85px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: transparent;
  color: var(--text-muted);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.tab-btn:hover {
  border-color: var(--accent-blue);
  color: var(--text-primary);
}

.tab-btn.active {
  background: var(--accent-blue);
  border-color: var(--accent-blue);
  color: white;
}

.chart-container {
  flex: 1;
  min-height: 250px;
}

/* 右侧控制区 */
.quick-controls {
  width: 250px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.control-group {
  background: var(--bg-panel);
  border-radius: 12px;
  padding: 12px;
  border: 1px solid var(--border-color);
}

.group-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 10px 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.action-buttons {
  display: flex;
  gap: 3px;
}

.action-btn {
  height: 50px;
  width: 80%;
  padding: 10px;
  border-radius: 8px;
  font-size: 20px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: all 0.2s;
}

.action-btn.start {
  background: var(--success);
  color: white;
}

.action-btn.start:hover:not(:disabled) {
  background: #059669;
}

.action-btn.stop {
  background: var(--danger);
  color: white;
}

.action-btn.stop:hover:not(:disabled) {
  background: var(--danger-hover);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.speed-input-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 12px;
}

.speed-input-row label {
  font-size: 12px;
  color: var(--text-muted);
}

.speed-input {
  padding: 8px 10px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 13px;
}

.jog-buttons {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.jog-btn {
  width: 80%;
  height: 40px;
  padding: 14px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: all 0.2s;
}

.jog-btn.up {
  background: rgba(16, 185, 129, 0.1);
  color: var(--success);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.jog-btn.up:hover {
  background: rgba(16, 185, 129, 0.2);
}

.jog-btn.halt {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.jog-btn.halt:hover {
  background: rgba(245, 158, 11, 0.2);
}

.jog-btn.down {
  background: rgba(239, 68, 68, 0.1);
  color: var(--danger);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.jog-btn.down:hover {
  background: rgba(239, 68, 68, 0.2);
}

.data-buttons {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.data-btn {
  width: 80%;
  padding: 12px;
  background: var(--bg-card);
  color: var(--text-regular);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: all 0.2s;
}

.data-btn:hover {
  border-color: var(--accent-blue);
  color: var(--text-primary);
}

/* 模态框 */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #1e293b;
  border-radius: 10px;
  width: 480px;
  max-width: 90vw;
  border: 1px solid #334155;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #334155;
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
  color: #f8fafc;
  display: flex;
  align-items: center;
  gap: 8px;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(255,255,255,0.05);
  color: var(--text-primary);
}

.modal-body {
  padding: 20px;
}

.device-panel {
  background: #1e293b;
  border-radius: 8px;
  padding: 16px;
  border: 1px solid #334155;
}

.device-panel + .device-panel {
  margin-top: 16px;
}

.panel-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 500;
  color: #f8fafc;
}

.refresh-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background: rgba(255,255,255,0.05);
  color: var(--text-primary);
}

.refresh-btn:disabled {
  cursor: not-allowed;
}

.refresh-btn .ri-refresh-line.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.device-select {
  width: 100%;
  padding: 12px;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  color: #f8fafc;
  margin-bottom: 12px;
  outline: none;
  font-size: 14px;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%23f8fafc' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
}

.device-select option {
  background: var(--bg-sidebar);
  color: var(--text-primary);
  padding: 8px 12px;
}

.connect-btn {
  width: 90%;
  height: 50px;
  padding: 12px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.connect-btn:hover:not(:disabled) {
  background: #2563eb;
}

.connect-btn:disabled {
  background: #64748b;
  cursor: not-allowed;
}

.connect-btn.connected {
  background: #10b981;
}

.connect-btn.connected:hover:not(:disabled) {
  background: #059669;
}

/* 项目设置模态框 */
.modal-container {
  background: #1e293b;
  border-radius: 10px;
  border: 1px solid #334155;
  overflow: hidden;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
}

.project-modal {
  /* width: 80%;
  height: 90%; */
  display: flex;
  flex-direction: column;
}

.project-modal .modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #334155;
  flex-shrink: 0;
}

.project-modal .modal-header h2 {
  margin: 0;
  font-size: 18px;
  color: #f8fafc;
}

.project-modal .modal-body {
  flex: 1;
  padding: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.project-modal .close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 6px;
  border-radius: 4px;
  transition: all 0.2s;
  font-size: 18px;
}

.project-modal .close-btn:hover {
  background: rgba(255,255,255,0.05);
  color: var(--text-primary);
}
</style>
