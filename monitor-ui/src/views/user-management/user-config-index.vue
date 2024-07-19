<template>
  <div class=" ">
    <Title :title="$t('m_title_setUp')"></Title>
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
        {
          label: 'm_field_personalInformation',
          path: '/userConfigIndex/userInformationModify'
        },
        {
          label: 'm_field_userManagement',
          path: '/userConfigIndex/userManagement'
        },
        {
          label: 'm_field_roleManagement',
          path: '/userConfigIndex/roleManagement'
        }
      ]
    }
  },
  mounted() {
    this.activeTab = this.$route.path
  },
  methods: {
    changeTab(path) {
      this.activeTab = path
      if (this.$route.path === path) {
        return
      }
      this.$router.push({path})
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
</style>
