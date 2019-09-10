<template>
  <div class="">
    <h3>{{title}}</h3>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
    }
  },
  props:{
    title: String
  }
}
</script>

<style scoped lang="less">
  h3 {
    padding: 10px 0;  
  }
</style>
