package com.luopingtech.ebike.rider.ui.h5

/**
 * 原生「后退」优先走 UniApp 页面栈，而不是 WebView.goBack。
 *
 * Uni H5 的 navigateTo 与浏览器 history 经常不同步：有时 canGoBack 为 false
 * 却仍有多层页面，有时 goBack 会落到 launch/profile 残留记录被误判关容器。
 * 问 getCurrentPages() 与产品预期一致，也消除「有时能回、有时直接退出」的抖动。
 */
internal object H5HostBackJs {
    val SCRIPT: String = """
        (function(){
          try {
            if (typeof getCurrentPages === 'function'
                && typeof uni !== 'undefined'
                && typeof uni.navigateBack === 'function') {
              var pages = getCurrentPages() || [];
              if (pages.length > 1) {
                uni.navigateBack({
                  delta: 1,
                  animationType: 'none',
                  animationDuration: 0
                });
                return 'kept';
              }
              return 'close';
            }
          } catch (e) {}
          return 'fallback';
        })();
    """.trimIndent()

    /** Android/iOS evaluateJavascript 回调里的 JSON 字符串 → kept|close|fallback */
    fun parseResult(raw: String?): String {
        if (raw.isNullOrBlank() || raw == "null") return "fallback"
        return raw.trim().removeSurrounding("\"").trim()
    }
}

/**
 * 注入平台标记：藏 Uni 标题栏、关页转场，避免与原生顶栏叠两层。
 */
internal object H5HostChromeJs {
    fun bootstrap(platform: String): String = """
        window.__riderNativePlatform='$platform';
        window.__riderNativeQueue=window.__riderNativeQueue||[];
        (function(){
          try {
            document.documentElement.classList.add('rider-native-host');
            var s = document.getElementById('rider-native-no-anim');
            if (!s) {
              s = document.createElement('style');
              s.id = 'rider-native-no-anim';
              document.head.appendChild(s);
            }
            s.textContent = ''
              + '.rider-native-host uni-page-head,'
              + '.rider-native-host .uni-page-head{'
              + '  display:none!important;height:0!important;min-height:0!important;'
              + '  padding:0!important;margin:0!important;overflow:hidden!important;'
              + '  pointer-events:none!important;opacity:0!important;'
              + '}'
              + '.rider-native-host uni-page-head[uni-page-head-type=default] ~ uni-page-wrapper,'
              + '.rider-native-host uni-page-wrapper{height:100%!important;}'
              + '.rider-native-host .uni-page-head-btn,'
              + '.rider-native-host .uni-page-head-hd .uni-btn-icon,'
              + '.rider-native-host .nav-back{'
              + '  display:none!important;width:0!important;min-width:0!important;'
              + '  height:0!important;padding:0!important;margin:0!important;'
              + '  overflow:hidden!important;pointer-events:none!important;'
              + '}'
              // 个人中心：只收顶空白，保持固定高，避免 my-property 负 margin 盖住快捷文字
              + '.rider-native-host .pages .info{'
              + '  padding-top:24px!important;'
              + '}'
              + '.rider-native-host .pages .top_container{'
              + '  height:262px!important;min-height:262px!important;'
              + '}'
              + '.rider-native-host uni-page,'
              + '.rider-native-host .uni-page,'
              + '.rider-native-host .uni-page--open,'
              + '.rider-native-host .uni-page--close,'
              + '.rider-native-host .uni-page--show,'
              + '.rider-native-host .uni-page--hide,'
              + '.rider-native-host .uni-main{'
              + '  animation:none!important;transition:none!important;'
              + '  -webkit-transition:none!important;-webkit-animation:none!important;'
              + '}';
          } catch (e) {}
        })();
        (function(){
          // 旧版 H5（费用异议等）仍先调 uni.chooseImage，在 App WebView 里会静默无回调。
          // 这里把它改成走 capturePhoto（原生拍照+签名上传），success 里直接给 CDN URL。
          if (window.__riderChooseImagePatched) return;
          window.__riderChooseImagePatched = true;
          var seq = 0;
          var pending = {};
          function ensureResultHook(){
            if (window.__riderNativeOnResult && window.__riderNativeOnResult.__riderPatched) return;
            var prev = window.__riderNativeOnResult;
            var hooked = function(id, result){
              var item = pending[String(id)];
              if (item) {
                delete pending[String(id)];
                clearTimeout(item.timer);
                if (result && typeof result === 'object' && result.__error) {
                  item.reject(new Error(String(result.msg || 'nativeHost error')));
                } else {
                  item.resolve(result);
                }
                return;
              }
              if (typeof prev === 'function') prev(id, result);
            };
            hooked.__riderPatched = true;
            window.__riderNativeOnResult = hooked;
            var q = window.__riderNativeQueue || [];
            window.__riderNativeQueue = [];
            for (var i = 0; i < q.length; i++) {
              try { hooked(q[i].id, q[i].result); } catch (e) {}
            }
          }
          function invokeCapture(source){
            ensureResultHook();
            var id = String(++seq);
            var payload = JSON.stringify({ id: id, method: 'capturePhoto', args: { source: source || 'camera' } });
            return new Promise(function(resolve, reject){
              var w = window;
              var timer = setTimeout(function(){
                delete pending[id];
                reject(new Error('nativeHost timeout: capturePhoto'));
              }, 120000);
              pending[id] = { resolve: resolve, reject: reject, timer: timer };
              try {
                if (w.__riderNative && typeof w.__riderNative.invoke === 'function') {
                  w.__riderNative.invoke(payload);
                } else if (w.webkit && w.webkit.messageHandlers && w.webkit.messageHandlers.riderNative) {
                  w.webkit.messageHandlers.riderNative.postMessage(payload);
                } else {
                  delete pending[id];
                  clearTimeout(timer);
                  reject(new Error('nativeHost unavailable'));
                }
              } catch (e) {
                delete pending[id];
                clearTimeout(timer);
                reject(e);
              }
            });
          }
          function hasNative(){
            var w = window;
            return !!(w.__riderNative || (w.webkit && w.webkit.messageHandlers && w.webkit.messageHandlers.riderNative)
              || w.__riderNativePlatform === 'android' || w.__riderNativePlatform === 'ios'
              || (document.documentElement && document.documentElement.classList.contains('rider-native-host')));
          }
          function patchChooseImage(){
            if (typeof uni === 'undefined' || typeof uni.chooseImage !== 'function') return false;
            if (uni.chooseImage.__riderPatched) return true;
            var orig = uni.chooseImage;
            uni.chooseImage = function(opts){
              opts = opts || {};
              if (!hasNative()) return orig.call(uni, opts);
              var sources = opts.sourceType || ['camera'];
              var source = (sources.indexOf('camera') >= 0) ? 'camera' : 'album';
              invokeCapture(source).then(function(res){
                var uri = String((res && res.uri) || '').trim();
                if (!uri) {
                  if (typeof opts.fail === 'function') opts.fail({ errMsg: 'chooseImage:fail empty uri' });
                  return;
                }
                if (typeof opts.success === 'function') {
                  opts.success({
                    tempFilePaths: [uri],
                    tempFiles: [{ path: uri, size: 1 }]
                  });
                }
              }).catch(function(err){
                if (typeof opts.fail === 'function') {
                  opts.fail({ errMsg: String((err && err.message) || 'chooseImage:fail') });
                }
              });
            };
            uni.chooseImage.__riderPatched = true;
            return true;
          }
          if (!patchChooseImage()) {
            var tries = 0;
            var timer = setInterval(function(){
              tries++;
              if (patchChooseImage() || tries > 80) clearInterval(timer);
            }, 250);
          }
        })();
    """.trimIndent()
}
