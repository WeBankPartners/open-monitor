<template>
  <Menu mode="horizontal" :theme="theme1" :active-name="activeName" @on-select="menuChange">
    <div class="logo" @click="routerChange">
      <img src="../assets/logo.png" />
      <span>{{$t('m_menu_systemName')}}</span>
    </div>
    <!--自定义看板-->
    <MenuItem name="viewConfigIndex">
      <i class="fa fa-gears" aria-hidden="true"></i>
      {{$t("m_custom_dashboard")}}
    </MenuItem>
    <!--对象看板-->
    <MenuItem name="endpointView">
      <Icon type="md-clipboard" :size="18" />
      {{$t("m_group_board")}}
    </MenuItem>
    <!--告警列表-->
    <MenuItem name="alarmManagement">
      <i class="fa fa-bell" aria-hidden="true"></i>
      {{$t("m_alarm_list")}}
    </MenuItem>
    <!--监控配置-->
    <MenuItem name="monitorConfigIndex">
      <Icon type="md-eye" :size="20"></Icon>
      {{$t("m_monitor_config")}}
    </MenuItem>
    <!--管理员配置-->
    <MenuItem name="adminConfig">
      <Icon type="md-person" :size="18"></Icon>
      {{$t("m_admin_config")}}
    </MenuItem>
    <div>
    </div>
    <!-- <div class="set-theme" :style="{background: !defaultTheme? 'white':'black'}" @click="changeTheme"></div> -->
    <div class="menu-right">
      <Dropdown trigger="click">
        <a href="javascript:void(0)">
          {{username}}
          <Icon type="ios-arrow-down"></Icon>
        </a>
        <DropdownMenu slot="list">
          <DropdownItem style="width:110px">
            <span @click="changeTheme">
              <a>{{$t('m_title_theme')}}:</a>
              <div class="set-theme" :style="{background: !defaultTheme ? 'white' : 'black'}" ></div>
            </span>
          </DropdownItem>
          <DropdownItem style="width:110px">
            <span @click="setUp">
              <a>{{$t('m_title_setUp')}}:</a>
              <div class="set-theme">
                <i class="fa fa-cog" aria-hidden="true"></i>
              </div>
            </span>
          </DropdownItem>
          <DropdownItem style="width:110px;">
            <span @click="logout">
              <a>{{$t('m_button_Logout')}}:</a>
              <div class="set-theme">
                <i class="fa fa-sign-out" aria-hidden="true"></i>
              </div>
            </span>
          </DropdownItem>
        </DropdownMenu>
      </Dropdown>
    </div>
    <div class="menu-right">
      <Dropdown @on-click="changeLang">
        <a href="javascript:void(0)">
          {{activeLang}}
          <Icon type="ios-arrow-down"></Icon>
        </a>
        <DropdownMenu slot="list">
          <template v-for="(langItem, langIndex) in lang">
            <DropdownItem
              :selected="langItem.label === activeLang"
              :name="langItem.label"
              :key="langIndex"
            >{{langItem.label}}</DropdownItem>
          </template>
        </DropdownMenu>
      </Dropdown>
    </div>
  </Menu>
</template>
<script>
import { getToken, removeToken} from '@/assets/js/cookies.ts'
import axios from 'axios'
import '@/assets/theme/dark/styls.less'
import '@/assets/theme/default/styls.less'
export default {
  data() {
    return {
      theme1: 'dark',
      defaultTheme: false,
      activeName: '',
      activeLang: '',
      langConfig: {
        'zh-CN': '中文',
        'en-US': 'English'
      },
      lang: [{
        label: '中文',
        value: 'zh-CN'
      }, {
        label: 'English',
        value: 'en-US'
      }],
      username: localStorage.getItem('username')
    }
  },
  created(){
    if (localStorage.getItem('theme')) {
      document.body.className = localStorage.getItem('theme')
      this.defaultTheme = false
    } else {
      document.body.className = ''
      this.defaultTheme = true
    }
    // document.body.className = localStorage.getItem('theme') ? localStorage.getItem('theme'): ''
  },
  mounted() {
    if (this.langConfig[localStorage.getItem('lang')] === undefined) {
      this.activeLang = this.langConfig[
        navigator.language || navigator.userLanguage
      ]
      this.setLocale(
        navigator.language || navigator.userLanguage === 'zh-CN'
          ? 'zh-CN'
          : 'en-US'
      )
    } else {
      this.activeLang = this.langConfig[localStorage.getItem('lang')]
    }
  },
  methods: {
    changeLang(name) {
      this.activeLang = name
      const lang = name === 'English' ? 'en-US' : 'zh-CN'
      this.setLocale(lang)
    },
    setLocale(lang) {
      localStorage.setItem('lang', lang)
      this.$i18n.locale = lang
      this.$validator.locale = lang
    },
    routerChange() {
      if (this.$route.name === 'dashboard') {
        return
      }
      this.$router.push({ name: 'dashboard' })
    },
    menuChange(name) {
      this.activeName = name
      if (this.$route.name === name) {
        return
      }
      this.$router.push({ name })
    },
    changeTheme() {
      const theme = localStorage.getItem('theme') ? '' : 'dark'
      localStorage.setItem('theme', theme)
      document.body.className = theme
      this.defaultTheme = !this.defaultTheme
    },
    setUp() {
      this.$router.push({path: '/userConfigIndex/userInformationModify'})
    },
    logout() {
      const url = require('../../src/assets/js/baseURL').baseURL_config + '/monitor/logout'
      axios({
        method: 'GET',
        url,
        headers: {
          'X-Auth-Token': getToken() || null
        }
      }).then(() => {
        removeToken()
        localStorage.removeItem('username')
        this.$router.push({path: 'login'})
      })
    }
  }
}
</script>
<style lang="less" scoped>
.ivu-menu-dark {
  background: #454a52;
}
.logo {
  float: left;
  height: inherit;
  padding-left: 30px;
  cursor: pointer;
  span {
    color: white;
    font-size: 16px;
    font-weight: bolder;
    vertical-align: top;
  }
  img {
    width: 40px;
    margin: 0 10px;
  }
}
.set-theme {
  float: right;
  width: 22px;
  height: 22px;
  border-radius: 4px;
  cursor: pointer;
  i {
    font-size: 18px;
  }
}
.menu-right {
  float: right;
  margin-right: 30px;
}
</style>
