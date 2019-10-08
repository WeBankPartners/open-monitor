<template>
  <div class="search-input-content">
    <Poptip placement="bottom" :width="parentConfig.poptipWidth">
      <input v-model.trim="ip.label"
      :placeholder="parentConfig.placeholder"
      @input="userInput"
      type="text"
      :style= "parentConfig.inputStyle"
      class="search-input" />
      <div class="poptip-content" slot="content" v-if="showSearchTips">
        <ul>
          <template v-for="(resItem, resIndex) in searchResult">
            <li class="ul-option" @click="choiceRes(resItem)" :key="resIndex">
              <Tag color="cyan" v-if="resItem.option_value.split(':')[1] == 'host'">host</Tag>
              <Tag color="blue" v-if="resItem.option_value.split(':')[1] == 'mysql'">mysql</Tag>
              <Tag color="geekblue" v-if="resItem.option_value.split(':')[1] == 'redis'">redis</Tag>
              <Tag color="purple" v-if="resItem.option_value.split(':')[1] == 'tomcat'">tomcat</Tag>
              <!-- <Tag color="default" v-else>{{resItem.option_value.split(':')[1]}}</Tag> -->
              <span>{{resItem.option_text}}</span>
            </li>
          </template>  
        </ul> 
      </div>
    </Poptip>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      ip: {
        label: '',
        value: '',
      },
      showSearchTips: false,
      searchResult: [],
    }
  },
  props: {
    parentConfig: Object
  },
  mounted(){
    if (Object.keys(this.$store.state.ip).length !== 0) {
      this.ip = this.$store.state.ip
    }
  },
  methods: {
    userInput () {
      this.showSearchTips = false
      this.request()
    },
    choiceRes (resItem) {
      this.ip.label = resItem.option_text
      this.ip.value = resItem.option_value
      this.ip.id = resItem.id
      this.ip.type = resItem.option_value.split(':')[1]
      this.$store.commit('storeip', this.ip)
      // this.$parent.getChartsConfig()
      this.showSearchTips = false
    },
    request () {
      if (!this.ip.label) {
        return
      }
      let searchParams = {
        search: this.ip.label
      }
      let params = Object.assign(searchParams, this.parentConfig.params)
      
      this.$httpRequestEntrance.httpRequestEntrance('GET',this.parentConfig.api, params, responseData => {
        this.searchResult = responseData
      })
      this.showSearchTips = true
    } 
  },
  components: {},
}
</script>

<style>
  .ivu-poptip-body {
    padding: 0;
  }
  .poptip-content {
    padding: 4px 0;
  }
  .poptip-content {
    max-height: 300px;
  }
</style>
<style scoped lang="less">
  .search-input {
    display: inline-block;
    height: 32px;
    padding: 4px 7px;
    font-size: 12px;
    border: 1px solid #dcdee2;
    border-radius: 4px;
    color: #515a6e;
    background-color: #fff;
    background-image: none;
    position: relative;
    cursor: text

  }
  .search-input:focus {
    outline: 0;
    border-color: #57a3f3;
  }

  .search-input-content {
    display: inline-block;
    vertical-align: middle; 
  }

  .ul-option {
    font-weight: 500;
    text-align: left;
    padding: 4px 16px;
    font-size: 12px;
    white-space: nowrap;
    cursor:pointer;
  }
  .ul-option:hover {
    background: @gray-hover;
  }
  .fa {
    padding-right: 8px;
  }
</style>
