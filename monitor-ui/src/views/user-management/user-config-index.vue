<template>
  <div class=" ">
     <Title :title="$t('title.setUp')"></Title>
    <section>
      <Tabs :value="activeTab" @on-click="changeTab"> 
        <template v-for="(tabItem, tabIndex) in tabs">
            <TabPane :label="$t(tabItem.label)" :name="tabItem.path" :key="tabIndex">
            </TabPane>
        </template>  
      </Tabs>
      <router-view></router-view>
    </section>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      activeTab: '/userConfigIndex/userInformationModify',
      tabs: [
        {label: 'field.personalInformation', path:'/userConfigIndex/userInformationModify'},
        {label: 'field.userManagement', path:'/userConfigIndex/userManagement'},
        {label: 'field.roleManagement', path:'/userConfigIndex/roleManagement'}
      ]
    }
  },
  mounted() {
    this.activeTab = this.$route.path
  },
  methods: {
    changeTab (path) {
      this.activeTab = path
      if (this.$route.path === path) return
      this.$router.push({path:path})
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
</style>
