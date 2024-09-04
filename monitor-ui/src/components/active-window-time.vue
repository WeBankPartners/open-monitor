<template>
  <div>
    <div class='mb-2' v-for="(time, index) in allTimeList" :key="index">
      <TimePicker
        style="width: 211px"
        v-model="allTimeList[index]"
        :clearable="false"
        format="HH:mm"
        :disabled="isDisabled"
        type="timerange"
        placement="bottom-end"
        @on-change="onTimePickerChange"
      >
      </TimePicker>
      <Button
        class="ml-2"
        type="error"
        ghost
        @click="deleteItem(index)"
        size="small"
        icon="md-trash"
        :disabled="isDisabled || allTimeList.length === 1"
      >
      </Button>
    </div>
    <Button
      class='add-button'
      type="success"
      :disabled="isDisabled"
      ghost
      @click="addItem()"
      size="small"
      icon="md-add"
    >
    </Button>
  </div>
</template>

<script>
import {isEmpty, cloneDeep} from 'lodash'

export default ({
  props: {
    activeWindowList: Array,
    isDisabled: Boolean
  },
  watch: {
    activeWindowList: {
      handler(list) {
        if (!isEmpty(list)) {
          this.allTimeList = []
          list.forEach(time => {
            const arr = time ? time.split('-') : ['00:00', '23:59']
            this.allTimeList.push(arr)
          })
        }
      },
      immediate: true
    }
  },
  data() {
    return {
      initialTime: ['00:00', '23:59'],
      allTimeList: [],
      finallTimeArr: []
    }
  },
  mounted(){
  },
  methods: {
    deleteItem(index) {
      this.allTimeList.splice(index, 1)
      this.onTimePickerChange()
    },
    addItem() {
      this.allTimeList.push(this.initialTime)
      this.onTimePickerChange()
    },
    onTimePickerChange() {
      this.finallTimeArr = []
      !isEmpty(this.allTimeList) && this.allTimeList.forEach(singleTime => {
        if (!isEmpty(singleTime) && singleTime.length === 2) {
          this.finallTimeArr.push(singleTime.join('-'))
        }
      })
      !isEmpty(this.finallTimeArr) && this.$emit('timeChange', cloneDeep(this.finallTimeArr))
    }
  }
})
</script>

<style scoped lang='less'>
.add-button {
  margin-left: 220px
}

</style>
