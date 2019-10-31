$.fn.extend({
	snow:function(obj){
		if(obj == undefined){
			var obj = new Object();
		}
		var settings = {
			snow:$(this),
			num:obj.num || 10,	//最多个数
			slow:obj.slow || 10,//最慢组度
			fast:obj.fast || 20,//最快速度
			conw:obj.conw || 1000, //布局宽度
			conh:obj.conh || 500, //布局高度
			minw:obj.minw || 10,//雪花最小宽度
			maxw:obj.maxw || 35,//雪花最大宽度
			addSpeed:obj.addSpeed || 800, //雪花增加的频率
			src:obj.src || 'images/img.png', //图片路径
			dir:obj.dir || 'ran' //下落方向 left right ran
		}
		settings.snow.width(settings.conw);//设置容器宽度
		//设置snow居中
		settings.snow.css({
			'margin-left':-settings.conw / 2,
			'left':'50%'
		})
		//图片加载完成后再启动
		function hasImg(){
			var img = new Image();
			img.src = settings.src;
			img.onload = function(){
				addStyle();//增加样式
				startImg();//增加雪花
			}
		}
		hasImg();
		//启动程序
		function startImg(){
			//定时器 开始增加雪花
			var timer = '';
				timer = setInterval(function(){
				var img = $('<img />');
				img.attr('src',settings.src);
				var ranTime = parseInt(Math.random() * (settings.fast - settings.slow)) + settings.slow;
				var cssText = 'left:' + parseInt(Math.random() * settings.conw) + 'px;width:' + (parseInt(Math.random() * (settings.maxw - settings.minw)) + settings.minw) + 'px;animation:mysnow ' + ranTime + 's infinite linear;-webkit-animation:mysnow ' + ranTime + 's infinite linear;';
				img.attr('style',cssText);
				settings.snow.append(img);
				if(settings.snow.find('img').length > settings.num){
					clearInterval(timer);
				}
			},settings.addSpeed);
		}
		//添加样式
		function addStyle(){
			var cssText = "";
			switch(settings.dir){
				case 'ran':
					cssText = moveRan([0,-20,30,0,-30,0]);//随机移动
					break;
				case 'right':
					cssText = moveRan([0.06*settings.conh,0.12*settings.conh,0.18*settings.conh,0.24*settings.conh,0.3*settings.conh,0.36*settings.conh]);//左移动
					break;
				case 'left':
					cssText = moveRan([-0.06*settings.conh,-0.12*settings.conh,-0.18*settings.conh,-0.24*settings.conh,-0.3*settings.conh,-0.36*settings.conh]);//右移动
					break;
				default:
					break;
			}
			var style = $('<style></style>');
				style.attr('id','snowStyle');
				style.append(cssText);
			$('head').append(style);
		}
		//随机移动
		function moveRan(arr){
			var cssText = ".snow {position: fixed;margin:0 auto;}.snow img {position: absolute;left:0;top:0;display: block;outline: none;}";
				cssText += "@keyframes mysnow {0%{transform:transform:translate("+arr[0]+"px,0px) rotate(0);opacity:1;}";
				cssText += "20%{transform:translate("+arr[1]+"px,"+0.2*settings.conh+"px) rotate(144deg);opacity:1;}";
				cssText += "40%{transform:translate("+arr[2]+"px,"+0.4*settings.conh+"px) rotate(288deg);}";
				cssText += "60%{transform:translate("+arr[3]+"px,"+0.6*settings.conh+"px) rotate(432deg);}";
				cssText += "80%{transform:translate("+arr[4]+"px,"+0.8*settings.conh+"px) rotate(576deg);opacity:1;}";
				cssText += "100%{transform:translate("+arr[5]+"px,"+settings.conh+"px) rotate(720deg);opacity:0;}}";
				cssText += "@-webkit-keyframes mysnow {0%{transform:transform:translate("+arr[0]+"px,0px) rotate(0);opacity:1;}";
				cssText += "20%{-webkit-transform:translate("+arr[1]+"px,"+0.2*settings.conh+"px) rotate(144deg);opacity:1;}";
				cssText += "40%{-webkit-transform:translate("+arr[2]+"px,"+0.4*settings.conh+"px) rotate(288deg);}";
				cssText += "60%{-webkit-transform:translate("+arr[3]+"px,"+0.6*settings.conh+"px) rotate(432deg);}";
				cssText += "80%{-webkit-transform:translate("+arr[4]+"px,"+0.8*settings.conh+"px) rotate(576deg);opacity:1;}";
				cssText += "100%{-webkit-transform:translate("+arr[5]+"px,"+settings.conh+"px) rotate(720deg);opacity:0;}}";
			return cssText;
		}
	}
})
//调用demo
/*
$('.snow').snow({
	num:20,	//最多个数
	slow:10,//最慢组度
	fast:20,//最快速度
	conw:310, //布局宽度
	conh:700, //布局高度
	minw:10,//雪花最小宽度
	maxw:35,//雪花最大宽度
	addSpeed:800, //雪花增加的频率
	src:'images/snow.png',//图片路径
	dir:'ran'//下落方向 left right ran
});
*/