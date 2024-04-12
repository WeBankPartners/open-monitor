function setLocalstorage (data) {
  const accessTokenObj = data.find(d => d.tokenType === 'accessToken')
  const refreshTokenObj = data.find(d => d.tokenType === 'refreshToken')
  localStorage.setItem('monitor-accessToken', accessTokenObj.token)
  localStorage.setItem('monitor-accessTokenExpirationTime', refreshTokenObj.expiration)
  localStorage.setItem('monitor-refreshToken', refreshTokenObj.token)
  localStorage.setItem('monitor-refreshTokenExpirationTime', refreshTokenObj.expiration)
}

function clearLocalstorage () {
  localStorage.removeItem('monitor-accessToken')
  localStorage.removeItem('monitor-accessTokenExpirationTime')
  localStorage.removeItem('monitor-refreshToken')
  localStorage.removeItem('monitor-refreshTokenExpirationTime')
}

export { setLocalstorage, clearLocalstorage }
