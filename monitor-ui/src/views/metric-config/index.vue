<template>
  <div class="monitor-metric-config">
    <Tabs v-model="activeName">
      <TabPane label="通用对象" name="1">
      </TabPane>
      <TabPane label="层级对象" name="2">
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