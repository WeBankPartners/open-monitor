<template>
  <div class=" ">
    <ul>
      <li v-for="(item, itemIndex) in data" class="tree-border" :key="itemIndex">
        <div class="tree-title" :style="stylePadding">
        <span>
          <strong>{{item.display_name}}</strong>
        </span>
        </div>
        <recursiveContainerItem
        :increment="count"
        v-if="item.children"
        :data="item.children"></recursiveContainerItem>
        <div>{{item.charts}}</div>
      </li>
    </ul>
  </div>
</template>

<script>
export default {
  name: 'recursiveContainerItem',
  data() {
    return {

    }
  },
  props:{
    data:{
      type:Array
    },
    increment:{
      type:Number,
      default:0
    }
  },
  computed:{
    count(){
      var c = this.increment;
      return ++c;
    },
    stylePadding(){
      return {
        'padding-left':this.count * 16 + 'px'
      }
    }
  },
  methods: {},
  components: {},
}
</script>

<style scoped lang="less">
  ul {
    padding: 0;
    margin: 0;
    list-style: none;
  }
 
  .tree-menu {
    height: 100%;
    padding: 0px 12px;
    border-right: 1px solid #e6e9f0;
  }

  .tree-menu-comm span {
    display: block;
    font-size: 12px;
    position: relative;
  }

  .tree-menu-comm span strong {
    display: block;
    width: 82%;
    position: relative;
    line-height: 22px;
    padding: 2px 0;
    padding-left: 5px;
    color: #161719;
    font-weight: normal;
  }

  .tree-title {
    border: 1px solid #fff;
    margin-top: 1px;
  }
  .tree-border {
    border: 1px solid #9966;
    padding: 8px;
    margin: 8px;;
  }
</style>
