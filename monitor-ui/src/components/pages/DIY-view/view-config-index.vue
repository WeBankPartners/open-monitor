<template>
  <div class="">
    <Select v-model="model1" style="width:200px">
      <Option v-for="item in cityList" :value="item.value" :key="item.value">{{ item.label }}</Option>
    </Select>
    <button class="btn btn-sm btn-confirm-f">
      <i class="fa fa-plus"></i>新建视图组
    </button>
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
          value: 'London',
          label: 'London'
        },
        {
          value: 'Sydney',
          label: 'Sydney'
        },
        {
          value: 'Ottawa',
          label: 'Ottawa'
        },
        {
          value: 'Paris',
          label: 'Paris'
        },
        {
          value: 'Canberra',
          label: 'Canberra'
        }
      ],
    model1: ''
    }
  },
  methods: {},
  components: {},
}
</script>

<style scoped lang="less">
</style>
