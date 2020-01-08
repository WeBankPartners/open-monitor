const setTitle = (title) => {
  document.title = title
}

export const cookies = {
  setAuthorization: value => {
    cookies.setCookie('Authorization', value)
  },
  getAuthorization: () => {
    return cookies.getCookie('Authorization')
  },
  deleteAuthorization: () => {
    cookies.delCookie('Authorization')
  },
  setUser: value => {
    cookies.setCookie('user', value)
  },
  getUser: () => {
    return cookies.getCookie('user')
  },
  setCookie: (name,value) => {
    const Days = 0.5
    const exp = new Date()
    exp.setTime(exp.getTime() + Days*24*60*60*1000)
    document.cookie = name + '=' + escape(value) + ';expires=' + exp.toGMTString()
  },
  getCookie: (name) => {
    const reg = new RegExp('(^| )' + name + '=([^;]*)(;|$)')
    const arr = document.cookie.match(reg)
    if (arr) {
      return unescape(arr[2])
    } else {
      return null
    }
  },
  delCookie: (name) => {
    const exp = new Date()
    exp.setTime(exp.getTime() - 1)
    const cval = cookies.getCookie(name)

    if(cval!=null)
      document.cookie= name + '=' + cval + ';expires=' + exp.toGMTString()
  }
}

export default {
  setTitle
}
