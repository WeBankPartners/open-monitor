<template>
  <div class="monitor-metric-config">
    <Tabs v-model="activeName">
      <!--通用对象-->
      <TabPane :label="$t('m_general_type')" name="1">
      </TabPane>
      <!--层级对象-->
      <TabPane :label="$t('field.resourceLevel')" name="2">
      </TabPane>
    </Tabs>
    <GeneralGroup v-if="activeName === '1'"></GeneralGroup>
    <LevelGroup v-if="activeName === '2'"></LevelGroup>
  </div>
</template>

<script>
import GeneralGroup from './general-group.vue'
import LevelGroup from './level-group.vue'
export default {
  components: {
    GeneralGroup,
    LevelGroup
  },
  data () {
    return {
      activeName: '1'
    }
  }
}
</script>

<style lang="less" scoped>
.monitor-metric-config {
  width: 100%;
}
</style>