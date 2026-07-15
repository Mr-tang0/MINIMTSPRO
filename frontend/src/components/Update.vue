<template>
  <div class="update-page">
    <div class="logo-container">
      <img src="/wails.png" class="update-logo" alt="Logo" />
    </div>

    <div class="content-container">
      <h2>发现新版本 {{ updateInfo.tagName }}</h2>
      <p class="description">当前版本: {{ updateInfo.currentVersion }}</p>
      <p class="description">点击更新获取更好体验！</p>

      <div class="release-notes" v-if="updateInfo.body">
        <h3>更新内容</h3>
        <pre>{{ updateInfo.body }}</pre>
      </div>

      <div class="update-actions">
        <button class="btn btn-primary" @click="handleUpdate">更新</button>
        <button class="btn btn-secondary" @click="handleCancel">取消</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import * as UpdateService from "../../bindings/MINIMTSPRO/backend/updateservice";
import { Browser, Window } from '@wailsio/runtime'


const updateInfo = ref({
  tagName: '无',
  currentVersion: '未知',
  htmlUrl: '',
  browserDownloadUrl: '',
  body: ''
})

onMounted(async () => {
    try{
        const release = await UpdateService.GetCachedRelease()

        if (release) {
            updateInfo.value.tagName = release.tag_name || release.TagName || '无'
            updateInfo.value.currentVersion = release.current_version || release.CurrentVersion || '未知'
            updateInfo.value.htmlUrl = release.html_url || release.HTMLURL || ''
            updateInfo.value.browserDownloadUrl = release.browser_download_url || release.BrowserDownloadURL || ''
            updateInfo.value.body = release.body || release.Body || ''
            
            if (!updateInfo.value.browserDownloadUrl && release.assets && release.assets.length > 0) {
                updateInfo.value.browserDownloadUrl = release.assets[0].browser_download_url
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

const handleUpdate = () => {
  const url = updateInfo.value.browserDownloadUrl || updateInfo.value.htmlUrl
  if (url) {
    Browser.OpenURL(url) 
    
    try {
      Window.Close() 
    } catch(e) {
      window.close()
    }
  }
}

const handleCancel = () => {
  try {
    Window.Close()
  } catch(e) {
    window.close()
  }
}


</script>

<style scoped>
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
  z-index: 9999;
}

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

.content-container {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  text-align: left;
  max-width: 500px;
}

h2 {
  color: #1e293b;
  font-size: 1.8rem;
  margin: 0 0 12px 0;
  font-weight: 700;
}

.description {
  color: #64748b;
  font-size: 1.05rem;
  margin: 0 0 16px 0;
}

.release-notes {
  width: 100%;
  background: #f8fafc;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
  border: 1px solid #e2e8f0;
  max-height: 200px;
  overflow-y: auto;
}

.release-notes h3 {
  color: #334155;
  font-size: 1rem;
  margin: 0 0 12px 0;
  font-weight: 600;
}

.release-notes pre {
  color: #475569;
  font-size: 0.9rem;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  font-family: inherit;
}

.update-actions {
  display: flex;
  gap: 16px;
  width: 100%;
  max-width: 320px;
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

.btn-primary {
  background: #3b82f6;
  color: white;
  box-shadow: 0 4px 6px -1px rgba(59, 130, 246, 0.2);
}

.btn-primary:hover {
  background: #2563eb;
}

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