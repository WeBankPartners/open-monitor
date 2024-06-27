<template>
  <div class="monitor-metric-config">
    <Tabs v-model="activeName">
      <!--通用对象-->
      <TabPane :label="$t('m_base_group')" name="1">
      </TabPane>
      <!--层级对象-->
      <TabPane :label="$t('m_field_resourceLevel')" name="2">
      </TabPane>
      <!--对象组-->
      <TabPane :label="$t('m_object_group')" name="3">
      </TabPane>
      <!--对象-->
      <TabPane :label="$t('m_object_design')+'('+$t('m_ready_only')+')'" name="4">
      </TabPane>
    </Tabs>
    <GeneralGroup v-if="activeName === '1'"></GeneralGroup>
    <LevelGroup v-if="activeName === '2'"></LevelGroup>
    <ObjectGroup v-if="activeName === '3'"></ObjectGroup>
    <Object v-if="activeName === '4'"></Object>
  </div>
</template>

<script>
import GeneralGroup from './general-group.vue'
import LevelGroup from './level-group.vue'
import ObjectGroup from './object-group.vue'
import Object from './object.vue'
export default {
  components: {
    GeneralGroup,
    LevelGroup,
    ObjectGroup,
    Object
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