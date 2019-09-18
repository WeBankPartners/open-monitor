import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import '@/assets/css/local.bootstrap.css'
import 'font-awesome/css/font-awesome.css'
import './plugins/iview.js'
import httpRequestEntrance from '@/assets/js/httpRequestEntrance.js'
import jquery from 'jquery'
import {validate} from '@/assets/js/validate.js'
import {tableUtil} from '@/assets/js/tableUtil.js'

Vue.prototype.$httpRequestEntrance = httpRequestEntrance
Vue.prototype.JQ = jquery
Vue.prototype.$validate = validate
Vue.prototype.$tableUtil = tableUtil

import PageTable from '@/components/components/table-page/page'
Vue.component('PageTable', PageTable)

Vue.config.productionTip = false


new Vue({
  render: h => h(App),
  store,
  router
}).$mount('#app')
