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
import Title from '@/components/components/Title'
export default {
  name: '',
  data() {
    return {
      activeTab: '/monitorConfigIndex/objectManagement',
      tabs: [
        {label: 'field.endpoint', path:'/monitorConfigIndex/objectManagement'},
        {label: 'field.group', path:'/monitorConfigIndex/groupManagement'},
        {label: 'field.threshold', path:'/monitorConfigIndex/thresholdManagement'},
        {label: 'field.log', path:'/monitorConfigIndex/logManagement'}
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