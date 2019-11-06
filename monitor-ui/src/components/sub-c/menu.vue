<template>
  <Menu mode="horizontal" :theme="theme1" :active-name="activeName" @on-select="menuChange">
    <div class="logo" @click="routerChange">
      <img src="../../assets/logo.png" />
      <span>{{$t('menu.systemName')}}</span>
    </div>
    <Submenu name>
      <template slot="title">
        <i class="fa fa-line-chart" aria-hidden="true"></i>
        {{$t("menu.view")}}
      </template>
      <MenuItem name="mainView">{{$t("menu.endpointView")}}</MenuItem>
      <MenuItem name="metricConfig">{{$t("menu.metricConfiguration")}}</MenuItem>
      <MenuItem name="viewConfigIndex">{{$t("menu.customViews")}}</MenuItem>
    </Submenu>
    <MenuItem name="monitorConfigIndex">
      <i class="fa fa-gears" aria-hidden="true"></i>
      {{$t("menu.configuration")}}
    </MenuItem>
    <MenuItem name="alarmManagement">
      <i class="fa fa-bell" aria-hidden="true"></i>
      {{$t("menu.alert")}}
    </MenuItem>
    <div style="float:right;padding-right:80px">
      <Dropdown @on-click="changeLang">
        <a href="javascript:void(0)">
          {{activeLang}}
          <Icon type="ios-arrow-down"></Icon>
        </a>
        <DropdownMenu slot="list">
          <template v-for="(langItem, langIndex) in lang">
            <DropdownItem
              :selected="langItem.label===activeLang"
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
export default {
  data() {
    return {
      theme1: "dark",
      activeName: "",
      activeLang: "",
      langConfig: {
        "zh-CN": "中文",
        "en-US": "En"
      },
      lang: [{ label: "中文", value: "zh-CN" }, { label: "En", value: "en-US" }]
    };
  },
  mounted() {
    if (this.langConfig[localStorage.getItem("lang")] === undefined) {
      this.activeLang = this.langConfig[
        navigator.language || navigator.userLanguage
      ];
      this.setLocale(
        navigator.language || navigator.userLanguage === "zh-CN"
          ? "zh-CN"
          : "en-US"
      );
    } else {
      this.activeLang = this.langConfig[localStorage.getItem("lang")];
    }
  },
  methods: {
    changeLang(name) {
      this.activeLang = name;
      let lang = name === "En" ? "en-US" : "zh-CN";
      this.setLocale(lang);
    },
    setLocale(lang) {
      localStorage.setItem("lang", lang);
      this.$i18n.locale = lang;
      this.$validator.locale = lang;
    },
    routerChange() {
      if (this.$route.name === "portal") return;
      this.$router.push({ name: "portal" });
    },
    menuChange(name) {
      this.activeName = name;
      if (this.$route.name === name) return;
      this.$router.push({ name: name });
    }
  }
};
</script>
<style lang="less" scoped>
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
    margin: 10px 20px 0;
  }
}
</style>
