<template>
  <div>
    <!-- {{ alertSoundTriggerOpen }} -->
    <!-- <button @click="changeAudioPlay">Start Audio</button> -->
    <audio id="alarmAudioPlay" src="../assets/alarm-audio/level1.mp3" loop></audio>
  </div>
</template>

<script>
import dayjs from 'dayjs'
export default {
  data() {
    return {
      alertSoundTriggerOpen: false,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      latestAlert: {}
    }
  },
  props: {
    // 数据集中包含的告警级别
    priority: {
      type: Array,
      default: () => ['medium', 'high']
    },
    // 查询频率，默认10s，在自定义看板中以具体刷新频率为准
    timeInterval: {
      type: Number,
      default: 10
    }
  },
  mounted(){
    this.getAlarm()
    this.audio = document.getElementById('alarmAudioPlay')
    this.interval = setInterval(() => {
      this.getAlarm()
    }, this.timeInterval * 1000)
    this.$once('hook:beforeDestroy', () => {
      clearInterval(this.interval)
    })
  },
  methods: {
    changeAudioPlay(trigger) {
      this.alertSoundTriggerOpen = trigger
      this.latestAlert = {}
      if (this.alertSoundTriggerOpen) {
        this.audio.volume = 0
        if (this.audio) {
          console.error('开启')
          this.audio.pause()
        }
      }
    },
    getAlarm() {
      if (!this.alertSoundTriggerOpen) {
        return
      }
      this.audio&&this.audio.pause()
      const now = dayjs().format('YYYY-MM-DD HH:mm:ss')
      const params = {
        page: {
          startIndex: 1,
          pageSize: 10
        },
        priority: this.priority
      }
      this.request(
        'POST',
        '/monitor/api/v1/alarm/problem/page',
        params,
        responseData => {
          const alertData = responseData.data || []
          if (alertData.length > 0) {
            if (this.latestAlert.id === alertData[0].id) {
              console.error('重复了')
              return
            }
            if (!this.priority.includes(alertData[0].s_priority)) {
              console.error('不在范围内')
              return
            }

            const latestAlert = alertData[0]
            if (this.latestAlert) {
              // 计算两个时间之间的总间隔
              const diff = dayjs(now).diff(dayjs(latestAlert.start_string))

              if (diff/1000 > this.timeInterval) {
                console.error('时间超期了')
                this.audio.pause()
                return
              }
            }
            if (this.latestAlert.id !== alertData[0].id) {
              this.latestAlert = alertData[0]
            }

            const priority = this.latestAlert.s_priority
            let iconSrc = ''
            if (priority === 'high') {
              iconSrc = require('../assets/img/icon_alarm_H_cube.png')
            }
            else if (priority === 'medium') {
              iconSrc = require('../assets/img/icon_alarm_M_cube.png')
            }
            else if (priority === 'low') {
              iconSrc = require('../assets/img/icon_alarm_L_cube.png')
            }
            const priorityI18n = {
              high: {
                text: this.$t('m_high'),
                color: '#da4e2b'
              },
              medium: {
                text: this.$t('m_medium'),
                color: '#f19d38'
              },
              low: {
                text: this.$t('m_low'),
                color: '#6fd16e'
              }
            }
            const priorityColor = {
              color: priorityI18n[priority].color,
              margin: '0 4px'
            }
            this.$Notice.open({
              duration: 5,
              render: () => (
                <div>
                  <div>
                    <img src={iconSrc} />
                    <span style={priorityColor}>[{priorityI18n[priority].text}]</span>
                    <div style="display:inline-block;width:230px;overflow: hidden;white-space: nowrap;text-overflow: ellipsis;line-height: 24px;vertical-align: bottom;">
                      {this.$t('m_new_alert')}:{this.latestAlert.alarm_name}
                    </div>
                  </div>
                  <div style="margin: 4px 0 0 30px;color: #808695;font-size:12px">
                    <span>{ this.$t('m_update_time') }:{ now }</span>
                  </div>
                </div>
              )
            })
            this.audio.play()
            this.audio.volume = 1
          }
        },
        {isNeedloading: false},
        () => {
          //
        }
      )
    }
  }
}
</script>
<style scoped lang="less">
</style>
