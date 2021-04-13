import Cookies from 'js-cookie'

// User
const tokenKey = 'Authorization'
export const getToken = () => Cookies.get(tokenKey)
const key = 'accessToken'
export const getPlatFormToken = () =>  {
  // eslint-disable-next-line no-useless-escape
  const reg = new RegExp('(?:(?:^|.*;\\s*)' + key + '\\s*\\=\\s*([^;]*).*$)|^.*$')
  return document.cookie.replace(reg, '$1')
}
export const setToken = (token: string) => Cookies.set(tokenKey, token)
export const removeToken = () => Cookies.remove(tokenKey)
