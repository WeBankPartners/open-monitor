const path = require('path')
const webpack = require('webpack')

module.exports = {
	assetsDir: 'wecube-monitor',
	chainWebpack: config => {
		const types = ['vue-modules', 'vue', 'normal-modules', 'normal']
		types.forEach(type => addStyleResource(config.module.rule('less').oneOf(type)))
	},
	css: {
		loaderOptions: {
			less: {
				javascriptEnabled: true
			}
		}
	},
	configureWebpack: {
		plugins: [
			new webpack.ProvidePlugin({
				$:"jquery",
				jQuery:"jquery",
				"windows.jQuery":"jquery"
			})
		]
  }
}
function addStyleResource(rule) {
	rule.use('style-resource')
	.loader('style-resources-loader')
	.options({
		patterns: [
			// 需要全局导入的less路径，
			path.resolve(__dirname, './src/assets/css/common.less')
		],
	})
}