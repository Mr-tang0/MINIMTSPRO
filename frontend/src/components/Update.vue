<template>
  <div class="update-page">
    <!-- 左侧 Logo 区域 -->
    <div class="logo-container">
      <img src="/wails.png" class="update-logo" alt="Logo" />
    </div>

    <!-- 右侧内容与操作区域 -->
    <div class="content-container">
      <h2>发现新版本 {{ updateInfo.tagName }}</h2>
      <p class="description">点击更新获取更好体验！</p>

      <div class="update-actions">
        <button class="btn btn-primary" @click="handleUpdate">更新</button>
        <button class="btn btn-secondary" @click="handleCancel">取消</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { GetCachedRelease } from "../../bindings/changeme/backend/updateservice";
import { Browser, Window } from '@wailsio/runtime'


const updateInfo = ref({
  tagName: '无',
  htmlUrl: 'https://github.com'
})

onMounted(async () => {
    try{
        // 调用我们在 Go 后端写好的获取缓存信息的方法
        const release = await GetCachedRelease()

        if (release) {
            updateInfo.value.tagName = release.tag_name
            updateInfo.value.htmlUrl = release.html_url
            if (release.assets.length > 0) {
                updateInfo.value.htmlUrl = release.assets[0].browser_download_url
            }
            
        }
        else{
            console.log("No cached release found")
        }
    }catch (error) {
        console.error("获取更新数据失败:", error)
        updateInfo.value.tagName = '版本获取失败'
    }

})

// 点击更新按钮
const handleUpdate = () => {
  if (updateInfo.value.htmlUrl) {
    Browser.OpenURL(updateInfo.value.htmlUrl) 
    
    try {
      Window.Close() 
    } catch(e) {
      window.close()
    }
  }
}

// 点击取消按钮
const handleCancel = () => {
  try {
    Window.Close() // Wails v3 原生关闭当前窗口
  } catch(e) {
    window.close() // 兜底降级方案
  }
}


</script>

<style scoped>
/* 完整填充界面，使用浅色背景渐变，横向弹性布局 */
.update-page {
  position: fixed;
  width: 100vw; 
  height: 100vh; 
  top: 0;
  left: 0;
  margin: 0;
  box-sizing: border-box;
  background: linear-gradient(135deg, #ffffff 0%, #f1f5f9 100%);
  display: flex;
  flex-direction: row; 
  align-items: center;
  justify-content: center;
  gap: 60px; 
  padding: 40px;
  font-family: 'Inter', -apple-system, system-ui, sans-serif;
  z-index: 9999;        /* 确保它在最上层 */
}

/* Logo 样式调整，适应浅色背景 */
.logo-container {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.update-logo {
  width: 120px;
  height: 120px;
  border-radius: 24px;
  background: #ffffff;
  padding: 20px;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.05), 0 8px 10px -6px rgba(0, 0, 0, 0.01);
  border: 1px solid #e2e8f0;
}

/* 右侧内容区，左对齐 */
.content-container {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  text-align: left;
}

/* 浅色主题的文字样式 */
h2 {
  color: #1e293b; /* 深色文字强调 */
  font-size: 1.8rem;
  margin: 0 0 12px 0;
  font-weight: 700;
}

.description {
  color: #64748b; /* 灰色次要文字 */
  font-size: 1.05rem;
  margin: 0 0 32px 0;
}

/* 按钮组样式 */
.update-actions {
  display: flex;
  gap: 16px;
  width: 100%;
  max-width: 320px; /* 限制按钮最大宽度使之美观 */
}

.btn {
    width: 100px;
    height: 35px;
  flex: 1;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  justify-content: center;
  align-items: center;
}

.btn:active {
  transform: scale(0.98);
}

/* 主要操作：更新按钮 */
.btn-primary {
  background: #3b82f6;
  color: white;
  box-shadow: 0 4px 6px -1px rgba(59, 130, 246, 0.2);
}

.btn-primary:hover {
  background: #2563eb;
}

/* 次要操作：取消按钮 (适配浅色主题) */
.btn-secondary {
  background: #ffffff;
  color: #475569;
  border: 1px solid #cbd5e1;
}

.btn-secondary:hover {
  background: #f8fafc;
  color: #334155;
  border-color: #94a3b8;
}
</style>