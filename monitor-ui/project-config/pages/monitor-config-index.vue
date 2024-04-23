<template>
  <div class="monitor-config-index">
    <Title :title="$t('menu.configuration')"></Title>
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
      activeTab: '/monitorConfigIndex/endpointManagement',
      tabs: [
        {label: 'field.endpoint', path:'/monitorConfigIndex/endpointManagement'},
        {label: 'field.resourceLevel', path:'/monitorConfigIndex/resourceLevel'},
        {label: 'field.group', path:'/monitorConfigIndex/groupManagement'},
        {label: 'field.threshold', path:'/monitorConfigIndex/thresholdManagement'},
        {label: 'field.log', path:'/monitorConfigIndex/logManagement'},
        {label: 'm_field_exporter', path:'/monitorConfigIndex/exporter'},
        {label: 'm_field_businessMonitor', path:'/monitorConfigIndex/businessMonitor'},
        {label: 'm_log_template', path:'/monitorConfigIndex/logTemplate'}
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
  },
}
</script>

<style scoped lang="less">
</style>
