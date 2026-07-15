<template>
  <div class="login-page">
    <!-- 背景图片 -->
    <div class="background-image"></div>
    
    <!-- 右侧毛玻璃卡片内容区 -->
    <div class="login-card">
      <!-- 圆形头像 -->
      <div class="avatar-container">
        <img v-if="!isRegister" src="../res/user.png" class="avatar" alt="用户头像" />
        <h2 class="welcome-title">{{ isRegister ? '创建新账号' : '欢迎回来' }}</h2>
      </div>

      <!-- 登录表单 -->
      <div v-if="!isRegister" class="form-container">
        <div class="input-group">
          <label>用户名</label>
          <input 
            v-model="username" 
            type="text" 
            placeholder="请输入用户名"
            @keyup.enter="handleLogin"
          />
        </div>

        <div class="input-group">
          <label>密  码</label>
          <input 
            v-model="password" 
            type="password" 
            placeholder="请输入密码"
            @keyup.enter="handleLogin"
          />
        </div>

        <!-- 登录按钮 -->
        <button class="submit-btn" @click="handleLogin">登 录</button>

        <!-- 底部链接 -->
        <div class="form-footer">
          <a href="#" class="switch-link" @click.prevent="toggleRegister">
            没有账号？立即注册
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="9 18 15 12 9 6"></polyline>
            </svg>
          </a>
        </div>
      </div>

      <!-- 注册表单 -->
      <div v-else class="form-container">
        <div class="input-group">
          <label>用户名</label>
          <input 
            v-model="registerForm.username" 
            type="text" 
            placeholder="请输入用户名"
          />
        </div>

        <div class="input-group">
          <label>邮箱</label>
          <input 
            v-model="registerForm.email" 
            type="email" 
            placeholder="请输入邮箱"
          />
        </div>

        <div class="input-group">
          <label>密码</label>
          <input 
            v-model="registerForm.password" 
            type="password" 
            placeholder="请输入密码"
          />
        </div>

        <div class="input-group">
          <label>确认密码</label>
          <input 
            v-model="registerForm.confirmPassword" 
            type="password" 
            placeholder="请再次输入密码"
          />
        </div>

        <!-- 注册按钮 -->
        <button class="submit-btn" @click="handleRegister">注 册</button>

        <!-- 底部链接 -->
        <div class="form-footer">
          <a href="#" class="switch-link" @click.prevent="toggleRegister">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="15 18 9 12 15 6"></polyline>
            </svg>
            已有账号？返回登录
          </a>
        </div>
      </div>

      <!-- 错误提示 -->
      <Transition name="fade">
        <p v-if="errorMsg" class="error-msg">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
          {{ errorMsg }}
        </p>
      </Transition>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import * as LoginService from '../../bindings/MINIMTSPRO/backend/loginservice'
import * as AppService from '../../bindings/MINIMTSPRO/backend/appservice'
import { Window } from '@wailsio/runtime'

const isRegister = ref(false)
const errorMsg = ref('')

const username = ref('')
const password = ref('')

const registerForm = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const toggleRegister = () => {
  isRegister.value = !isRegister.value
  errorMsg.value = ''
}

const validatePassword = (value) => {
  if (value.length < 6) {
    errorMsg.value = '密码长度不能少于6位'
    return false
  }
  return true
}

const getAuthError = (user) => user?.app_json?.error || user?.AppJson?.error || ''

const handleLogin = async () => {
  if (!username.value || !password.value) {
    errorMsg.value = '请填写完整的登录信息'
    return
  }
  if (!(username.value === 'admin' && password.value === 'admin') && !validatePassword(password.value)) return

  try {
    errorMsg.value = ''
    const user = await LoginService.Login(username.value, password.value)
    const authError = getAuthError(user)
    if (authError) {
      errorMsg.value = authError
      return
    }
    if (!user?.name && !user?.id) {
      errorMsg.value = '用户名或密码错误'
      return
    }
    await AppService.CallMINIMTSWindow()
    try {
      Window.Close()
    } catch(e) {
      window.close()
    }
  } catch (error) {
    errorMsg.value = error?.message || String(error) || '登录失败'
  }
}

const handleRegister = async () => {
  if (!registerForm.username || !registerForm.email || !registerForm.password || !registerForm.confirmPassword) {
    errorMsg.value = '请填写完整的注册信息'
    return
  }
  if (!validatePassword(registerForm.password)) return
  if (registerForm.password !== registerForm.confirmPassword) {
    errorMsg.value = '两次输入的密码不一致'
    return
  }

  try {
    errorMsg.value = ''
    const user = await LoginService.Login(`__register__:${JSON.stringify({ username: registerForm.username, email: registerForm.email })}`, registerForm.password)
    const authError = getAuthError(user)
    if (authError) {
      errorMsg.value = authError
      return
    }
    username.value = registerForm.username
    password.value = ''
    toggleRegister()
  } catch (error) {
    errorMsg.value = error?.message || String(error) || '注册失败'
  }
}

onMounted(async () => {
  try {
    const info = await LoginService.Login('__last_login__', '000000')
    if (info?.username) {
      username.value = info.username
    }
  } catch (error) {
    console.error(error)
  }
})
</script>

<style scoped>
/* 页面基础布局 */
.login-page {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 100%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: flex-end; /* 保持靠右布局 */
  padding-right: 5%; /* 使用百分比，适配不同宽度的屏幕 */
  font-family: 'Inter', -apple-system, system-ui, sans-serif;
}

.background-image {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: url('../res/loginWin.jpg') no-repeat center center;
  background-size: cover;
  z-index: 1;
}

/* 核心：毛玻璃卡片 */
.login-card {
  position: relative;
  z-index: 10;
  width: 340px;
  min-height: 500px;
  background: rgba(15, 23, 42, 0.6); /* 半透明深色背景 */
  backdrop-filter: blur(16px); /* 强毛玻璃效果 */
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.1); /* 极细的高光边框 */
  border-radius: 15px;
  padding: 40px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5); /* 柔和深邃的阴影 */
  display: flex;
  flex-direction: column;
}

/* 头像与标题区 */
.avatar-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 24px; /* 稍微缩小下边距使整体紧凑 */
}

.avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  border: 3px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.2);
  object-fit: cover;
  background-color: #f1f5f9;
}

.welcome-title {
  margin: 16px 0 0 0;
  color: #ffffff;
  font-size: 1.25rem;
  font-weight: 600;
  letter-spacing: 1px;
}

/* 表单容器 */
.form-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
    align-items: center;
}

.input-group {
  display: flex;
  flex-direction: row; /* 横向排列 */
  align-items: center; /* 垂直居中对齐 */
  gap: 12px; /* 标签和输入框的间距 */
}

.input-group label {
  color: #cbd5e1;
  font-size: 0.9rem;
  font-weight: 500;
  width: 70px; /* 固定宽度以对齐所有输入框 */
  text-align: right; /* 右对齐标签，更整齐 */
  flex-shrink: 0;
}

.input-group input {
  flex: 1; /* 占据剩余全部宽度 */
  box-sizing: border-box;
  padding: 10px 14px; /* 稍微减小上下内边距使得整体更加紧凑 */
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.05); /* 半透明底色，比透明更好看 */
  color: #ffffff;
  font-size: 0.95rem;
  outline: none;
  transition: all 0.3s ease;
}

.input-group input:focus {
  border-color: #3b82f6;
  background: rgba(255, 255, 255, 0.1);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

.input-group input::placeholder {
  color: #64748b;
}

/* 提交按钮 (实心主色调，更强视觉引导) */
.submit-btn {
  width: 60%;
  height: 50px;
  padding: 12px;
  border: none;
  border-radius: 8px;
  background: #3b82f6; /* 主题蓝 */
  color: #ffffff;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-top: 8px;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.submit-btn:hover {
  background: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(59, 130, 246, 0.4);
}

.submit-btn:active {
  transform: translateY(1px);
}

/* 底部链接区 */
.form-footer {
  margin-top: 8px;
  display: flex;
  justify-content: center;
}

.switch-link {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #94a3b8;
  font-size: 0.85rem;
  text-decoration: none;
  transition: color 0.2s ease;
}

.switch-link:hover {
  color: #ffffff;
}

/* 错误提示 */
.error-msg {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  color: #fca5a5;
  font-size: 0.85rem;
  padding: 10px;
  border-radius: 8px;
  margin-top: 16px;
  text-align: center;
}

/* Vue 错误提示过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-5px);
}
</style>
