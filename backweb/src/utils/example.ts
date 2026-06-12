import { http } from './request'
import { authAPI, userAPI, imageAPI, adminAPI } from '../api'

// 使用示例
export const apiExamples = {
  // 登录示例
  async loginExample() {
    try {
      const response = await authAPI.login('admin@slink.org', 'password123')
      console.log('登录成功:', response.data)
      return response.data
    } catch (error) {
      console.error('登录失败:', error)
      throw error
    }
  },

  // 注册示例
  async registerExample() {
    try {
      const response = await authAPI.register(
        'newuser@example.com',
        'password123',
        '新用户',
        '123456'
      )
      console.log('注册成功:', response.data)
      return response.data
    } catch (error) {
      console.error('注册失败:', error)
      throw error
    }
  },

  // 发送验证码示例
  async sendVerificationCodeExample() {
    try {
      const response = await authAPI.sendVerificationCode('user@example.com')
      console.log('验证码发送成功:', response.data)
      return response.data
    } catch (error) {
      console.error('验证码发送失败:', error)
      throw error
    }
  },

  // 上传图片示例
  async uploadImageExample(file: File) {
    try {
      const response = await imageAPI.uploadImage(file, (progress) => {
        console.log(`上传进度: ${progress}%`)
      })
      console.log('图片上传成功:', response.data)
      return response.data
    } catch (error) {
      console.error('图片上传失败:', error)
      throw error
    }
  },

  // 批量上传图片示例
  async uploadMultipleImagesExample(files: File[]) {
    try {
      const response = await imageAPI.uploadMultipleImages(files, (progress) => {
        console.log(`批量上传进度: ${progress}%`)
      })
      console.log('批量上传成功:', response.data)
      return response.data
    } catch (error) {
      console.error('批量上传失败:', error)
      throw error
    }
  },

  // 获取用户图片列表示例
  async getUserImagesExample() {
    try {
      const response = await imageAPI.getImageList(1, 20)
      console.log('用户图片列表:', response.data)
      return response.data
    } catch (error) {
      console.error('获取图片列表失败:', error)
      throw error
    }
  },

  // 删除图片示例
  async deleteImageExample(imageId: number) {
    try {
      const response = await imageAPI.deleteImage(imageId)
      console.log('图片删除成功:', response.data)
      return response.data
    } catch (error) {
      console.error('图片删除失败:', error)
      throw error
    }
  },

  // 管理员获取所有用户示例
  async adminGetAllUsersExample() {
    try {
      const response = await adminAPI.getAllUsers()
      console.log('所有用户列表:', response.data)
      return response.data
    } catch (error) {
      console.error('获取用户列表失败:', error)
      throw error
    }
  },

  // 管理员删除用户示例
  async adminDeleteUserExample(userId: number) {
    try {
      const response = await adminAPI.deleteUser(userId)
      console.log('用户删除成功:', response.data)
      return response.data
    } catch (error) {
      console.error('用户删除失败:', error)
      throw error
    }
  },

  // 更新用户权限示例
  async updateUserRoleExample(userId: number, isAdmin: boolean) {
    try {
      const response = await adminAPI.updateUserRole(userId, isAdmin)
      console.log('用户权限更新成功:', response.data)
      return response.data
    } catch (error) {
      console.error('用户权限更新失败:', error)
      throw error
    }
  },

  // 获取系统统计信息示例
  async getSystemStatsExample() {
    try {
      const response = await adminAPI.getSystemStats()
      console.log('系统统计信息:', response.data)
      return response.data
    } catch (error) {
      console.error('获取系统统计失败:', error)
      throw error
    }
  },

  // 直接使用http方法的示例
  async directHttpExample() {
    try {
      // GET请求
      const getResponse = await http.get('/api/test')
      console.log('GET响应:', getResponse.data)

      // POST请求
      const postResponse = await http.post('/api/test', { data: 'test' })
      console.log('POST响应:', postResponse.data)

      // PUT请求
      const putResponse = await http.put('/api/test/1', { data: 'updated' })
      console.log('PUT响应:', putResponse.data)

      // DELETE请求
      const deleteResponse = await http.delete('/api/test/1')
      console.log('DELETE响应:', deleteResponse.data)

      // 文件上传
      const file = new File(['test'], 'test.txt', { type: 'text/plain' })
      const uploadResponse = await http.upload('/api/upload', file, (progress) => {
        console.log(`上传进度: ${progress}%`)
      })
      console.log('上传响应:', uploadResponse.data)

    } catch (error) {
      console.error('HTTP请求失败:', error)
      throw error
    }
  }
}

export default apiExamples 