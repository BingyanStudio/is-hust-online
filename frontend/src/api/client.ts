import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
})

api.interceptors.response.use(
  (res) => {
    if (res.data.code !== 0) {
      return Promise.reject(new Error(res.data.message || 'Request failed'))
    }
    return res.data.data
  },
  (error) => {
    return Promise.reject(error)
  },
)

export default api
