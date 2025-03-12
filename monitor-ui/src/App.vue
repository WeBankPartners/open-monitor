<template>
  <div id="app" class="c-dark">
    <router-view />
  </div>
</template>

<script>
export default {
  name: 'app'
}
</script>

<style>
@import "./assets/css/global.css";
#app {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  line-height: inherit;
  font-size: 12px;

  width: 100%;
  height: 100%;
}

html,body{
  width: 100%;
  height: 100%;
  /* overflow: hidden; */
  margin: 0;
  padding: 0;
}

</style>
