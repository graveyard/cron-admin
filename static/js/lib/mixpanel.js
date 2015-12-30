/* begin mixpanel section */
(function(f,b){if(!b.__SV){var a,e,i,g;window.mixpanel=b;b._i=[];b.init=function(a,e,d){function f(b,h){var a=h.split(".");2==a.length&&(b=b[a[0]],h=a[1]);b[h]=function(){b.push([h].concat(Array.prototype.slice.call(arguments,0)))}}var c=b;"undefined"!==typeof d?c=b[d]=[]:d="mixpanel";c.people=c.people||[];c.toString=function(b){var a="mixpanel";"mixpanel"!==d&&(a+="."+d);b||(a+=" (stub)");return a};c.people.toString=function(){return c.toString(1)+".people (stub)"};i="disable track track_pageview track_links track_forms register register_once alias unregister identify name_tag set_config people.set people.set_once people.increment people.append people.union people.track_charge people.clear_charges people.delete_user".split(" ");
for(g=0;g<i.length;g++)f(c,i[g]);b._i.push([a,e,d])};b.__SV=1.2;a=f.createElement("script");a.type="text/javascript";a.async=!0;a.src="undefined"!==typeof MIXPANEL_CUSTOM_LIB_URL?MIXPANEL_CUSTOM_LIB_URL:"//cdn.mxpnl.com/libs/mixpanel-2-latest.min.js";e=f.getElementsByTagName("script")[0];e.parentNode.insertBefore(a,e)}})(document,window.mixpanel||[]);

mixpanel.init("d51d878eb79f67d9b9c91ba45c713b9f");
/* end mixpanel section */

var MixpanelMixin = {
  sendToMixpanel: function(eventName, props) {
    //console.log("event: ", eventName, " otherProps: ", props);
    mixpanel.track(eventName, _.extend({"section": this.props.section, "page": this.constructor.displayName}, props));
  }
};

// send along the props - only concern is around if someone were
// to put PII in the properties, but really no PII should be touched
// by react code
var MixpanelMountMixin = {
  componentDidMount: function() {
    this.sendToMixpanel("load", this.props);
  }
};

var MixpanelUpdateMixin = {
  componentDidUpdate: function(prevProps, prevState) {
    this.sendToMixpanel("update", this.props);
  }
};
