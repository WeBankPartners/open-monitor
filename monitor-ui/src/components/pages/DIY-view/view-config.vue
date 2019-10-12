<template>
  <div class=" ">
      <header>
        <div style="display:flex;justify-content:space-between">
            <div class="header-name">
                <i class="fa fa-th-large fa-2x" aria-hidden="true"></i>
                <span>   {{$route.params.label}}</span>
            </div>
            <div class="header-tools"> 
                <i class="fa fa-plus-square-o fa-2x" @click="addItem" aria-hidden="true"></i>
            </div>
        </div>
      </header>
      <grid-layout
        :layout.sync="layoutData"
        :col-num="12"
        :row-height="30"
        :is-draggable="true"
        :is-resizable="true"
        :is-mirrored="false"
        :vertical-compact="true"

        :use-css-transforms="true">
      <grid-item v-for="(item,index) in layoutData"
                   :x="item.x"
                   :y="item.y"
                   :w="item.w"
                   :h="item.h"
                   :i="item.i"
                   :key="index">
        <div style="display:flex;justify-content:flex-end">
          <div class="header-grid header-grid-name">
            {{item.i}}
          </div>
          <div class="header-grid header-grid-tools"> 
            <i class="fa fa-eye" aria-hidden="true"></i>
            <i class="fa fa-cog" @click="editGrid(item)" aria-hidden="true"></i>
            <i class="fa fa-trash" @click="removeGrid(item)" aria-hidden="true"></i>
          </div>
        </div>
        </grid-item>
      </grid-layout>
      {{layoutData}}
  </div>
</template>

<script>
import VueGridLayout from 'vue-grid-layout';
export default {
  name: '',
  data() {
    return {
      layoutData: [
        //   {'x':0,'y':0,'w':2,'h':2,'i':'0'},
        //   {'x':1,'y':1,'w':2,'h':2,'i':'1'},
      ]
    }
  },
  mounted() {
    if(this.$validate.isEmpty_reset(this.$route.params)) {
      this.$router.push({path:'viewConfigIndex'})
    } 
  },
  methods: {
    addItem() {
      const key = ((new Date()).valueOf()).toString().substring(10)
      let item = {"x":0,"y":0,"w":6,"h":7,
                  "i": `default${key}`,
                  id: key
                }
      this.layoutData.push(item);
    },
    editGrid(item) {
      this.$router.push({name: 'editView', params:item})
    },
    removeGrid(item) {
      this.layoutData.splice(this.layoutData.indexOf(item), 1);
    },
  },
  components: {
    GridLayout: VueGridLayout.GridLayout,
    GridItem: VueGridLayout.GridItem,
  },
}
</script>

<style scoped lang="less">
  .header-grid {
    flex-grow: 1;
    text-align: end;
    line-height: 32px;
    i {
      margin: 0 4px;
      cursor: pointer;
      // border: 1px solid red;
      // font-size: 14px
    } 
  }
</style>
<style scoped lang="less">


.columns {
    -moz-columns: 120px;
    -webkit-columns: 120px;
    columns: 120px;
}

.vue-grid-item:not(.vue-grid-placeholder) {
    background: @gray-f;
    border: 1px solid @gray-f;
}

.vue-grid-item.resizing {
    opacity: 0.9;
}

.vue-grid-item.static {
    background: #cce;
}

.vue-grid-item .text {
    font-size: 24px;
    text-align: center;
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    margin: auto;
    height: 100%;
    width: 100%;
}

.vue-grid-item .no-drag {
    height: 100%;
    width: 100%;
}

.vue-grid-item .minMax {
    font-size: 12px;
}

.vue-grid-item .add {
    cursor: pointer;
}

.vue-draggable-handle {
    position: absolute;
    width: 20px;
    height: 20px;
    top: 0;
    left: 0;
    background: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='10' height='10'><circle cx='5' cy='5' r='5' fill='#999999'/></svg>") no-repeat;
    background-position: bottom right;
    padding: 0 8px 8px 0;
    background-repeat: no-repeat;
    background-origin: content-box;
    box-sizing: border-box;
    cursor: pointer;
}

</style>
