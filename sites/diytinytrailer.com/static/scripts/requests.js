(function () {
  'use strict';
  
  var module = {
      options: [],
      header: [navigator.platform, navigator.userAgent, navigator.appVersion, navigator.vendor, window.opera],
      dataos: [
          { name: 'Windows Phone', value: 'Windows Phone', version: 'OS' },
          { name: 'Windows', value: 'Win', version: 'NT' },
          { name: 'iPhone', value: 'iPhone', version: 'OS' },
          { name: 'iPad', value: 'iPad', version: 'OS' },
          { name: 'Kindle', value: 'Silk', version: 'Silk' },
          { name: 'Android', value: 'Android', version: 'Android' },
          { name: 'PlayBook', value: 'PlayBook', version: 'OS' },
          { name: 'BlackBerry', value: 'BlackBerry', version: '/' },
          { name: 'Macintosh', value: 'Mac', version: 'OS X' },
          { name: 'Linux', value: 'Linux', version: 'rv' },
          { name: 'Palm', value: 'Palm', version: 'PalmOS' }
      ],
      databrowser: [
          { name: 'Chrome', value: 'Chrome', version: 'Chrome' },
          { name: 'Firefox', value: 'Firefox', version: 'Firefox' },
          { name: 'Safari', value: 'Safari', version: 'Version' },
          { name: 'Internet Explorer', value: 'MSIE', version: 'MSIE' },
          { name: 'Edge Chromium', value: 'Edge', version: 'Edge' },
          { name: 'Opera', value: 'Opera', version: 'Opera' },
          { name: 'BlackBerry', value: 'CLDC', version: 'CLDC' },
          { name: 'Mozilla', value: 'Mozilla', version: 'Mozilla' }
      ],
      init: function () {
          var agent = this.header.join(' '),
              os = this.matchItem(agent, this.dataos),
              browser = this.matchItem(agent, this.databrowser);
          
          return { os: os, browser: browser };
      },
      matchItem: function (string, data) {
          var i = 0,
              j = 0,
              html = '',
              regex,
              regexv,
              match,
              matches,
              version;
          
          for (i = 0; i < data.length; i += 1) {
              regex = new RegExp(data[i].value, 'i');
              match = regex.test(string);
              if (match) {
                  regexv = new RegExp(data[i].version + '[- /:;]([\\d._]+)', 'i');
                  matches = string.match(regexv);
                  version = '';
                  if (matches) { if (matches[1]) { matches = matches[1]; } }
                  if (matches) {
                      matches = matches.split(/[._]+/);
                      for (j = 0; j < matches.length; j += 1) {
                          if (j === 0) {
                              version += matches[j] + '.';
                          } else {
                              version += matches[j];
                          }
                      }
                  } else {
                      version = '0';
                  }
                  return {
                      name: data[i].name,
                      version: parseFloat(version)
                  };
              }
          }
          return { name: 'unknown', version: 0 };
      }
  };
  
  var e = module.init();

  function OSVersion(){
    return e.os.version;
  }
  function OS(){
    return e.os.name;
  }
  function NavUserAgent(){
    return navigator.userAgent;
  }
  function NavAppVersion(){
    return navigator.appVersion;
  }
  function NavPlatform(){
    return navigator.platform;
  }
  function NavVendor(){
    return navigator.vendor;
  }        
  function BrowserName() {
      return e.browser.name;
  }
  function BrowserVersion(){
    return e.browser.version;
  }

  window.OSVersion = OSVersion;
  window.OS = OS;
  window.NavUserAgent = NavUserAgent
  window.NavAppVersion = NavAppVersion
  window.NavPlatform = NavPlatform
  window.NavVendor = NavVendor
  window.BrowserName = BrowserName
  window.BrowserVersion = BrowserVersion
}());

// Get the session
function readCookie(name) {
  var nameEQ = name + "=";
  var ca = document.cookie.split(';');
  for(var i=0;i < ca.length;i++) {
      var c = ca[i];
      while (c.charAt(0)==' ') c = c.substring(1,c.length);
      if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
  }
  return null;
}

// Attempt at browser detection by functionality
var BrowserByFunction = function() {

  if (BrowserByFunction.prototype._cachedResult)
      return browser.prototype._cachedResult;
  // Opera 8.0+
  var isOpera = (!!window.opr && !!opr.addons) || !!window.opera || navigator.userAgent.indexOf(' OPR/') >= 0;
  
  // Firefox 1.0+
  var isFirefox = typeof InstallTrigger !== 'undefined';
  
  // Safari 3.0+ "[object HTMLElementConstructor]" 
  var isSafari = /constructor/i.test(window.HTMLElement) || (function (p) { return p.toString() === "[object SafariRemoteNotification]"; })(!window['safari'] || (typeof safari !== 'undefined' && safari.pushNotification));
  
  // Internet Explorer 6-11
  var isIE = /*@cc_on!@*/false || !!document.documentMode;
  
  // Edge 20+
  var isEdge = !isIE && !!window.StyleMedia;
  
  // Chrome 1 - 79
  var isChrome = !!window.chrome && (!!window.chrome.webstore || !!window.chrome.runtime);
  
  // Edge (based on chromium) detection
  var isEdgeChromium = isChrome && (navigator.userAgent.indexOf("Edg") != -1);
  
  // Blink engine detection
  var isBlink = (isChrome || isOpera) && !!window.CSS;
  
  return (BrowserByFunction.prototype._cachedResult =
      isOpera ? 'Opera' :
      isFirefox ? 'Firefox' :
      isSafari ? 'Safari' :
      isChrome ? 'Chrome' :
      isIE ? 'IE' :
      isEdge ? 'Edge' :
      isEdgeChromium ? 'EdgeChromium' :
      isOpera ? 'Opera' :
      isBlink ? 'Blink' :
      "unknown");
}

$( document ).ready(function() {

  bdata = {
      'sessionid': readCookie('session_token'),
      'functionalbrowser': BrowserByFunction(),
      'osversion': OSVersion() + "",
      'os' : OS(),
      'useragent' : NavUserAgent(),
      'navappversion': NavAppVersion(), 
      'navplatform': NavPlatform(),
      'navbrowser': BrowserName(),
      'browserversion': BrowserVersion() + "",
      'ptag': ptag
  }

  $.ajax({
      url: '/request/api/v1',
      type: 'post',
      contentType: 'application/json; charset=utf-8',
      data: JSON.stringify(bdata),
      dataType: 'json',
      success : function(r) {
          console.log( r.status + ' ' +  r.message);
      },
      failure: function(errMsg) {
          console.log(errMsg);
      },
      error: function(errMsg) {
          console.log(errMsg);
      }
  });
}); 