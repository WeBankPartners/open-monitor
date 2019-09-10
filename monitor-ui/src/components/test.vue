<template>
  <div class="hello">
    <h1>TEST</h1>
    <Button type="primary" @click="test">Primary</Button>
  </div>
</template>

<script>
export default {
  name: 'TEST',
  methods: {
    test () {
      // let params = {
      //   endpoint: ['192.168.0.16'],
      //   metric: ['cpu.used.percent'],
      //   time: '-1800'
      // }
      // this.$httpRequestEntrance.httpRequestEntrance('POST','/api/v1/dashboard/chart', params, responseData => {
      //   console.log(responseData) 
      // })

      // let params = {
      //   search: 1
      // }
      // this.$httpRequestEntrance.httpRequestEntrance('GET','/api/v1/dashboard/search', params, responseData => {
      //   console.log(responseData) 
      // })

    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h3 {
  margin: 40px 0 0;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
