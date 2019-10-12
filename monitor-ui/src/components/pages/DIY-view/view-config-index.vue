<template>
  <div class="">
    <Select v-model="model1" style="width:200px">
      <Option v-for="item in cityList" :value="item.value" :key="item.value">{{ item.label }}</Option>
    </Select>
    <button class="btn btn-sm btn-confirm-f">
      <i class="fa fa-plus"></i>新建视图组
    </button>

    <section>
      <ul>
        <template v-for="(panalItem,panalIndex) in dataList">
          <li :key="panalIndex" class="panalItem" @click="goToPanal(panalItem)">{{panalItem.label}}</li>
        </template>
      </ul>
    </section>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      cityList: [
        {
          value: 'New York',
          label: 'New York'
        },
        {
          value: 'Paris',
          label: 'Paris'
        }
      ],
      model1: '',
      dataList: [
        {label:'模板一' ,value: 'v1'},
        {label:'模板二' ,value: 'v2'},
        {label:'模板三' ,value: 'v3'}
      ]
    }
  },
  methods: {
    goToPanal(panalItem) {
      this.$router.push({name:'viewConfig',params:panalItem})
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
.panalItem {
  border: 1px solid @blue-2;
  margin: 4px;
  padding: 16px;
  display:inline-block;
}
</style>
