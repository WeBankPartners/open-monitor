<template>
  <div class="page" id="endpointView">
    <!-- <Title :title="$t('m_menu_endpointView')"></Title> -->
    <EndpointViewComponent  ref="endpointViewComponentRef"></EndpointViewComponent>
  </div>
</template>
<script>
import EndpointViewComponent from '@/components/endpoint-view-component'
export default {
  name: 'endpoint-view',
  data() {
    return {
    }
  },
  mounted() {
  },
  methods: {
    // #endregion
  },
  components: {
    EndpointViewComponent
  }
}
</script>
