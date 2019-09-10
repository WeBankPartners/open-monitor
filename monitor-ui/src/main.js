import Vue from 'vue'
import App from './App.vue'
import router from './router'
import 'font-awesome/css/font-awesome.css'
import './plugins/iview.js'
import httpRequestEntrance from '@/assets/js/httpRequestEntrance.js'

Vue.config.productionTip = false


Vue.prototype.$httpRequestEntrance = httpRequestEntrance

new Vue({
  render: h => h(App),
  router
}).$mount('#app')
