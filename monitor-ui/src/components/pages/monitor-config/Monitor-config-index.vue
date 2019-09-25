<template>
  <div class="monitor-config-index">
    <Title title="监控管理主页"></Title>
    <section>
      <Tabs :value="activeTab" @on-click="changeTab"> 
        <template v-for="(tabItem, tabIndex) in tabs">
            <TabPane :label="tabItem.label" :name="tabItem.path" :key="tabIndex">
            </TabPane>
        </template>  
      </Tabs>
      <router-view></router-view>
    </section>
  </div>
</template>

<script>
import Title from '@/components/components/Title'
export default {
  name: '',
  data() {
    return {
      activeTab: '/monitorConfigIndex/objectManagement',
      tabs: [
        {label: '对象管理', path:'/monitorConfigIndex/objectManagement'},
        {label: '组管理', path:'/monitorConfigIndex/groupManagement'},
        {label: '阀值管理', path:'/monitorConfigIndex/thresholdManagement'},
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
  components: {
    Title
  },
}
</script>

<style scoped lang="less">
</style>