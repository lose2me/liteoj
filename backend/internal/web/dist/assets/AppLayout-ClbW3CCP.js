import{d as U,w as J,h as d,F as Te,T as ae,m as Fe,S as ie,v as ee,r as O,u as le,a as Me,b as He,c as Ne,o as Ie,e as We,f as B,i as de,g as Z,j as Pe,p as j,k as De,l as Le,n as Ue,q as i,s as Y,t as k,x as _,y as N,z as Ae,A as je,L as Xe,B as Ye,C as ce,D as Ve,E as qe,G as Ke,H as Ge,I,J as te,N as Je,K as Qe,M as ue,O as P,P as Q,Q as R,R as s,U as re,V as X,W as T,X as L,Y as F,Z as W,_ as he,$ as me,a0 as G,a1 as Ze,a2 as et,a3 as tt,a4 as rt}from"./index-GV3JmupW.js";import{h as ot}from"./http-ByXr2EJk.js";import{t as m}from"./index-DGOhw6Ge.js";import{u as nt}from"./use-message-D6Rvf_q7.js";import{N as oe}from"./Input-DcFON9dl.js";import{_ as fe}from"./_plugin-vue_export-helper-DlAUqK2U.js";import{N as st,a as at,b as it}from"./Menu-BDu7m828.js";import{u as ne,f as se}from"./get-CjhBgZP8.js";import"./Suffix-lvV7-qj5.js";import"./Dropdown-DNc5_MKS.js";import"./Popover-B_1KmcmA.js";import"./use-compitable-BO06rvls.js";const lt=U({name:"NDrawerContent",inheritAttrs:!1,props:{blockScroll:Boolean,show:{type:Boolean,default:void 0},displayDirective:{type:String,required:!0},placement:{type:String,required:!0},contentClass:String,contentStyle:[Object,String],nativeScrollbar:{type:Boolean,required:!0},scrollbarProps:Object,trapFocus:{type:Boolean,default:!0},autoFocus:{type:Boolean,default:!0},showMask:{type:[Boolean,String],required:!0},maxWidth:Number,maxHeight:Number,minWidth:Number,minHeight:Number,resizable:Boolean,onClickoutside:Function,onAfterLeave:Function,onAfterEnter:Function,onEsc:Function},setup(e){const t=O(!!e.show),r=O(null),u=de(Z);let f=0,h="",l=null;const b=O(!1),g=O(!1),S=B(()=>e.placement==="top"||e.placement==="bottom"),{mergedClsPrefixRef:p,mergedRtlRef:o}=le(e),w=Me("Drawer",o,p),z=n,y=a=>{g.value=!0,f=S.value?a.clientY:a.clientX,h=document.body.style.cursor,document.body.style.cursor=S.value?"ns-resize":"ew-resize",document.body.addEventListener("mousemove",E),document.body.addEventListener("mouseleave",z),document.body.addEventListener("mouseup",n)},M=()=>{l!==null&&(window.clearTimeout(l),l=null),g.value?b.value=!0:l=window.setTimeout(()=>{b.value=!0},300)},V=()=>{l!==null&&(window.clearTimeout(l),l=null),b.value=!1},{doUpdateHeight:q,doUpdateWidth:K}=u,D=a=>{const{maxWidth:v}=e;if(v&&a>v)return v;const{minWidth:$}=e;return $&&a<$?$:a},A=a=>{const{maxHeight:v}=e;if(v&&a>v)return v;const{minHeight:$}=e;return $&&a<$?$:a};function E(a){var v,$;if(g.value)if(S.value){let x=((v=r.value)===null||v===void 0?void 0:v.offsetHeight)||0;const H=f-a.clientY;x+=e.placement==="bottom"?H:-H,x=A(x),q(x),f=a.clientY}else{let x=(($=r.value)===null||$===void 0?void 0:$.offsetWidth)||0;const H=f-a.clientX;x+=e.placement==="right"?H:-H,x=D(x),K(x),f=a.clientX}}function n(){g.value&&(f=0,g.value=!1,document.body.style.cursor=h,document.body.removeEventListener("mousemove",E),document.body.removeEventListener("mouseup",n),document.body.removeEventListener("mouseleave",z))}He(()=>{e.show&&(t.value=!0)}),Ne(()=>e.show,a=>{a||n()}),Ie(()=>{n()});const c=B(()=>{const{show:a}=e,v=[[ee,a]];return e.showMask||v.push([Pe,e.onClickoutside,void 0,{capture:!0}]),v});function C(){var a;t.value=!1,(a=e.onAfterLeave)===null||a===void 0||a.call(e)}return We(B(()=>e.blockScroll&&t.value)),j(De,r),j(Le,null),j(Ue,null),{bodyRef:r,rtlEnabled:w,mergedClsPrefix:u.mergedClsPrefixRef,isMounted:u.isMountedRef,mergedTheme:u.mergedThemeRef,displayed:t,transitionName:B(()=>({right:"slide-in-from-right-transition",left:"slide-in-from-left-transition",top:"slide-in-from-top-transition",bottom:"slide-in-from-bottom-transition"})[e.placement]),handleAfterLeave:C,bodyDirectives:c,handleMousedownResizeTrigger:y,handleMouseenterResizeTrigger:M,handleMouseleaveResizeTrigger:V,isDragging:g,isHoverOnResizeTrigger:b}},render(){const{$slots:e,mergedClsPrefix:t}=this;return this.displayDirective==="show"||this.displayed||this.show?J(d("div",{role:"none"},d(Te,{disabled:!this.showMask||!this.trapFocus,active:this.show,autoFocus:this.autoFocus,onEsc:this.onEsc},{default:()=>d(ae,{name:this.transitionName,appear:this.isMounted,onAfterEnter:this.onAfterEnter,onAfterLeave:this.handleAfterLeave},{default:()=>J(d("div",Fe(this.$attrs,{role:"dialog",ref:"bodyRef","aria-modal":"true",class:[`${t}-drawer`,this.rtlEnabled&&`${t}-drawer--rtl`,`${t}-drawer--${this.placement}-placement`,this.isDragging&&`${t}-drawer--unselectable`,this.nativeScrollbar&&`${t}-drawer--native-scrollbar`]}),[this.resizable?d("div",{class:[`${t}-drawer__resize-trigger`,(this.isDragging||this.isHoverOnResizeTrigger)&&`${t}-drawer__resize-trigger--hover`],onMouseenter:this.handleMouseenterResizeTrigger,onMouseleave:this.handleMouseleaveResizeTrigger,onMousedown:this.handleMousedownResizeTrigger}):null,this.nativeScrollbar?d("div",{class:[`${t}-drawer-content-wrapper`,this.contentClass],style:this.contentStyle,role:"none"},e):d(ie,Object.assign({},this.scrollbarProps,{contentStyle:this.contentStyle,contentClass:[`${t}-drawer-content-wrapper`,this.contentClass],theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar}),e)]),this.bodyDirectives)})})),[[ee,this.displayDirective==="if"||this.displayed||this.show]]):null}}),{cubicBezierEaseIn:dt,cubicBezierEaseOut:ct}=Y;function ut({duration:e="0.3s",leaveDuration:t="0.2s",name:r="slide-in-from-bottom"}={}){return[i(`&.${r}-transition-leave-active`,{transition:`transform ${t} ${dt}`}),i(`&.${r}-transition-enter-active`,{transition:`transform ${e} ${ct}`}),i(`&.${r}-transition-enter-to`,{transform:"translateY(0)"}),i(`&.${r}-transition-enter-from`,{transform:"translateY(100%)"}),i(`&.${r}-transition-leave-from`,{transform:"translateY(0)"}),i(`&.${r}-transition-leave-to`,{transform:"translateY(100%)"})]}const{cubicBezierEaseIn:ht,cubicBezierEaseOut:mt}=Y;function ft({duration:e="0.3s",leaveDuration:t="0.2s",name:r="slide-in-from-left"}={}){return[i(`&.${r}-transition-leave-active`,{transition:`transform ${t} ${ht}`}),i(`&.${r}-transition-enter-active`,{transition:`transform ${e} ${mt}`}),i(`&.${r}-transition-enter-to`,{transform:"translateX(0)"}),i(`&.${r}-transition-enter-from`,{transform:"translateX(-100%)"}),i(`&.${r}-transition-leave-from`,{transform:"translateX(0)"}),i(`&.${r}-transition-leave-to`,{transform:"translateX(-100%)"})]}const{cubicBezierEaseIn:bt,cubicBezierEaseOut:gt}=Y;function vt({duration:e="0.3s",leaveDuration:t="0.2s",name:r="slide-in-from-right"}={}){return[i(`&.${r}-transition-leave-active`,{transition:`transform ${t} ${bt}`}),i(`&.${r}-transition-enter-active`,{transition:`transform ${e} ${gt}`}),i(`&.${r}-transition-enter-to`,{transform:"translateX(0)"}),i(`&.${r}-transition-enter-from`,{transform:"translateX(100%)"}),i(`&.${r}-transition-leave-from`,{transform:"translateX(0)"}),i(`&.${r}-transition-leave-to`,{transform:"translateX(100%)"})]}const{cubicBezierEaseIn:pt,cubicBezierEaseOut:wt}=Y;function yt({duration:e="0.3s",leaveDuration:t="0.2s",name:r="slide-in-from-top"}={}){return[i(`&.${r}-transition-leave-active`,{transition:`transform ${t} ${pt}`}),i(`&.${r}-transition-enter-active`,{transition:`transform ${e} ${wt}`}),i(`&.${r}-transition-enter-to`,{transform:"translateY(0)"}),i(`&.${r}-transition-enter-from`,{transform:"translateY(-100%)"}),i(`&.${r}-transition-leave-from`,{transform:"translateY(0)"}),i(`&.${r}-transition-leave-to`,{transform:"translateY(-100%)"})]}const Ct=i([k("drawer",`
 word-break: break-word;
 line-height: var(--n-line-height);
 position: absolute;
 pointer-events: all;
 box-shadow: var(--n-box-shadow);
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 background-color: var(--n-color);
 color: var(--n-text-color);
 box-sizing: border-box;
 `,[vt(),ft(),yt(),ut(),_("unselectable",`
 user-select: none; 
 -webkit-user-select: none;
 `),_("native-scrollbar",[k("drawer-content-wrapper",`
 overflow: auto;
 height: 100%;
 `)]),N("resize-trigger",`
 position: absolute;
 background-color: #0000;
 transition: background-color .3s var(--n-bezier);
 `,[_("hover",`
 background-color: var(--n-resize-trigger-color-hover);
 `)]),k("drawer-content-wrapper",`
 box-sizing: border-box;
 `),k("drawer-content",`
 height: 100%;
 display: flex;
 flex-direction: column;
 `,[_("native-scrollbar",[k("drawer-body-content-wrapper",`
 height: 100%;
 overflow: auto;
 `)]),k("drawer-body",`
 flex: 1 0 0;
 overflow: hidden;
 `),k("drawer-body-content-wrapper",`
 box-sizing: border-box;
 padding: var(--n-body-padding);
 `),k("drawer-header",`
 font-weight: var(--n-title-font-weight);
 line-height: 1;
 font-size: var(--n-title-font-size);
 color: var(--n-title-text-color);
 padding: var(--n-header-padding);
 transition: border .3s var(--n-bezier);
 border-bottom: 1px solid var(--n-divider-color);
 border-bottom: var(--n-header-border-bottom);
 display: flex;
 justify-content: space-between;
 align-items: center;
 `,[N("main",`
 flex: 1;
 `),N("close",`
 margin-left: 6px;
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 `)]),k("drawer-footer",`
 display: flex;
 justify-content: flex-end;
 border-top: var(--n-footer-border-top);
 transition: border .3s var(--n-bezier);
 padding: var(--n-footer-padding);
 `)]),_("right-placement",`
 top: 0;
 bottom: 0;
 right: 0;
 border-top-left-radius: var(--n-border-radius);
 border-bottom-left-radius: var(--n-border-radius);
 `,[N("resize-trigger",`
 width: 3px;
 height: 100%;
 top: 0;
 left: 0;
 transform: translateX(-1.5px);
 cursor: ew-resize;
 `)]),_("left-placement",`
 top: 0;
 bottom: 0;
 left: 0;
 border-top-right-radius: var(--n-border-radius);
 border-bottom-right-radius: var(--n-border-radius);
 `,[N("resize-trigger",`
 width: 3px;
 height: 100%;
 top: 0;
 right: 0;
 transform: translateX(1.5px);
 cursor: ew-resize;
 `)]),_("top-placement",`
 top: 0;
 left: 0;
 right: 0;
 border-bottom-left-radius: var(--n-border-radius);
 border-bottom-right-radius: var(--n-border-radius);
 `,[N("resize-trigger",`
 width: 100%;
 height: 3px;
 bottom: 0;
 left: 0;
 transform: translateY(1.5px);
 cursor: ns-resize;
 `)]),_("bottom-placement",`
 left: 0;
 bottom: 0;
 right: 0;
 border-top-left-radius: var(--n-border-radius);
 border-top-right-radius: var(--n-border-radius);
 `,[N("resize-trigger",`
 width: 100%;
 height: 3px;
 top: 0;
 left: 0;
 transform: translateY(-1.5px);
 cursor: ns-resize;
 `)])]),i("body",[i(">",[k("drawer-container",`
 position: fixed;
 `)])]),k("drawer-container",`
 position: relative;
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 pointer-events: none;
 `,[i("> *",`
 pointer-events: all;
 `)]),k("drawer-mask",`
 background-color: rgba(0, 0, 0, .3);
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `,[_("invisible",`
 background-color: rgba(0, 0, 0, 0)
 `),Ae({enterDuration:"0.2s",leaveDuration:"0.2s",enterCubicBezier:"var(--n-bezier-in)",leaveCubicBezier:"var(--n-bezier-out)"})])]),St=Object.assign(Object.assign({},ce.props),{show:Boolean,width:[Number,String],height:[Number,String],placement:{type:String,default:"right"},maskClosable:{type:Boolean,default:!0},showMask:{type:[Boolean,String],default:!0},to:[String,Object],displayDirective:{type:String,default:"if"},nativeScrollbar:{type:Boolean,default:!0},zIndex:Number,onMaskClick:Function,scrollbarProps:Object,contentClass:String,contentStyle:[Object,String],trapFocus:{type:Boolean,default:!0},onEsc:Function,autoFocus:{type:Boolean,default:!0},closeOnEsc:{type:Boolean,default:!0},blockScroll:{type:Boolean,default:!0},maxWidth:Number,maxHeight:Number,minWidth:Number,minHeight:Number,resizable:Boolean,defaultWidth:{type:[Number,String],default:251},defaultHeight:{type:[Number,String],default:251},onUpdateWidth:[Function,Array],onUpdateHeight:[Function,Array],"onUpdate:width":[Function,Array],"onUpdate:height":[Function,Array],"onUpdate:show":[Function,Array],onUpdateShow:[Function,Array],onAfterEnter:Function,onAfterLeave:Function,drawerStyle:[String,Object],drawerClass:String,target:null,onShow:Function,onHide:Function}),$t=U({name:"Drawer",inheritAttrs:!1,props:St,setup(e){const{mergedClsPrefixRef:t,namespaceRef:r,inlineThemeDisabled:u}=le(e),f=Ye(),h=ce("Drawer","-drawer",Ct,Ke,e,t),l=O(e.defaultWidth),b=O(e.defaultHeight),g=ne(te(e,"width"),l),S=ne(te(e,"height"),b),p=B(()=>{const{placement:n}=e;return n==="top"||n==="bottom"?"":se(g.value)}),o=B(()=>{const{placement:n}=e;return n==="left"||n==="right"?"":se(S.value)}),w=n=>{const{onUpdateWidth:c,"onUpdate:width":C}=e;c&&I(c,n),C&&I(C,n),l.value=n},z=n=>{const{onUpdateHeight:c,"onUpdate:width":C}=e;c&&I(c,n),C&&I(C,n),b.value=n},y=B(()=>[{width:p.value,height:o.value},e.drawerStyle||""]);function M(n){const{onMaskClick:c,maskClosable:C}=e;C&&D(!1),c&&c(n)}function V(n){M(n)}const q=Ve();function K(n){var c;(c=e.onEsc)===null||c===void 0||c.call(e),e.show&&e.closeOnEsc&&Ge(n)&&(q.value||D(!1))}function D(n){const{onHide:c,onUpdateShow:C,"onUpdate:show":a}=e;C&&I(C,n),a&&I(a,n),c&&!n&&I(c,n)}j(Z,{isMountedRef:f,mergedThemeRef:h,mergedClsPrefixRef:t,doUpdateShow:D,doUpdateHeight:z,doUpdateWidth:w});const A=B(()=>{const{common:{cubicBezierEaseInOut:n,cubicBezierEaseIn:c,cubicBezierEaseOut:C},self:{color:a,textColor:v,boxShadow:$,lineHeight:x,headerPadding:H,footerPadding:be,borderRadius:ge,bodyPadding:ve,titleFontSize:pe,titleTextColor:we,titleFontWeight:ye,headerBorderBottom:Ce,footerBorderTop:Se,closeIconColor:$e,closeIconColorHover:ze,closeIconColorPressed:ke,closeColorHover:xe,closeColorPressed:Be,closeIconSize:Re,closeSize:Ee,closeBorderRadius:_e,resizableTriggerColorHover:Oe}}=h.value;return{"--n-line-height":x,"--n-color":a,"--n-border-radius":ge,"--n-text-color":v,"--n-box-shadow":$,"--n-bezier":n,"--n-bezier-out":C,"--n-bezier-in":c,"--n-header-padding":H,"--n-body-padding":ve,"--n-footer-padding":be,"--n-title-text-color":we,"--n-title-font-size":pe,"--n-title-font-weight":ye,"--n-header-border-bottom":Ce,"--n-footer-border-top":Se,"--n-close-icon-color":$e,"--n-close-icon-color-hover":ze,"--n-close-icon-color-pressed":ke,"--n-close-size":Ee,"--n-close-color-hover":xe,"--n-close-color-pressed":Be,"--n-close-icon-size":Re,"--n-close-border-radius":_e,"--n-resize-trigger-color-hover":Oe}}),E=u?qe("drawer",void 0,A,e):void 0;return{mergedClsPrefix:t,namespace:r,mergedBodyStyle:y,handleOutsideClick:V,handleMaskClick:M,handleEsc:K,mergedTheme:h,cssVars:u?void 0:A,themeClass:E==null?void 0:E.themeClass,onRender:E==null?void 0:E.onRender,isMounted:f}},render(){const{mergedClsPrefix:e}=this;return d(Xe,{to:this.to,show:this.show},{default:()=>{var t;return(t=this.onRender)===null||t===void 0||t.call(this),J(d("div",{class:[`${e}-drawer-container`,this.namespace,this.themeClass],style:this.cssVars,role:"none"},this.showMask?d(ae,{name:"fade-in-transition",appear:this.isMounted},{default:()=>this.show?d("div",{"aria-hidden":!0,class:[`${e}-drawer-mask`,this.showMask==="transparent"&&`${e}-drawer-mask--invisible`],onClick:this.handleMaskClick}):null}):null,d(lt,Object.assign({},this.$attrs,{class:[this.drawerClass,this.$attrs.class],style:[this.mergedBodyStyle,this.$attrs.style],blockScroll:this.blockScroll,contentStyle:this.contentStyle,contentClass:this.contentClass,placement:this.placement,scrollbarProps:this.scrollbarProps,show:this.show,displayDirective:this.displayDirective,nativeScrollbar:this.nativeScrollbar,onAfterEnter:this.onAfterEnter,onAfterLeave:this.onAfterLeave,trapFocus:this.trapFocus,autoFocus:this.autoFocus,resizable:this.resizable,maxHeight:this.maxHeight,minHeight:this.minHeight,maxWidth:this.maxWidth,minWidth:this.minWidth,showMask:this.showMask,onEsc:this.handleEsc,onClickoutside:this.handleOutsideClick}),this.$slots)),[[je,{zIndex:this.zIndex,enabled:this.show}]])}})}}),zt={title:String,headerClass:String,headerStyle:[Object,String],footerClass:String,footerStyle:[Object,String],bodyClass:String,bodyStyle:[Object,String],bodyContentClass:String,bodyContentStyle:[Object,String],nativeScrollbar:{type:Boolean,default:!0},scrollbarProps:Object,closable:Boolean},kt=U({name:"DrawerContent",props:zt,slots:Object,setup(){const e=de(Z,null);e||Qe("drawer-content","`n-drawer-content` must be placed inside `n-drawer`.");const{doUpdateShow:t}=e;function r(){t(!1)}return{handleCloseClick:r,mergedTheme:e.mergedThemeRef,mergedClsPrefix:e.mergedClsPrefixRef}},render(){const{title:e,mergedClsPrefix:t,nativeScrollbar:r,mergedTheme:u,bodyClass:f,bodyStyle:h,bodyContentClass:l,bodyContentStyle:b,headerClass:g,headerStyle:S,footerClass:p,footerStyle:o,scrollbarProps:w,closable:z,$slots:y}=this;return d("div",{role:"none",class:[`${t}-drawer-content`,r&&`${t}-drawer-content--native-scrollbar`]},y.header||e||z?d("div",{class:[`${t}-drawer-header`,g],style:S,role:"none"},d("div",{class:`${t}-drawer-header__main`,role:"heading","aria-level":"1"},y.header!==void 0?y.header():e),z&&d(Je,{onClick:this.handleCloseClick,clsPrefix:t,class:`${t}-drawer-header__close`,absolute:!0})):null,r?d("div",{class:[`${t}-drawer-body`,f],style:h,role:"none"},d("div",{class:[`${t}-drawer-body-content-wrapper`,l],style:b,role:"none"},y)):d(ie,Object.assign({themeOverrides:u.peerOverrides.Scrollbar,theme:u.peers.Scrollbar},w,{class:`${t}-drawer-body`,contentClass:[`${t}-drawer-body-content-wrapper`,l],contentStyle:b}),y),y.footer?d("div",{class:[`${t}-drawer-footer`,p],style:o,role:"none"},y.footer()):null)}}),xt={class:"login-form"},Bt={key:0,class:"redirect-hint"},Rt=U({__name:"LoginCard",emits:["success"],setup(e,{emit:t}){const r=t,u=ue(),f=me(),h=nt(),l=O({username:"",password:""}),b=O(!1),g=async()=>{var S,p;if(!l.value.username||!l.value.password){h.warning(m.auth.needUserAndPwd);return}b.value=!0;try{const{data:o}=await ot.post("/auth/login",l.value);u.setAuth(o.token,o.user),h.success(m.auth.loginOk),r("success")}catch(o){h.error(((p=(S=o==null?void 0:o.response)==null?void 0:S.data)==null?void 0:p.error)||m.auth.loginFailed)}finally{b.value=!1}};return(S,p)=>(P(),Q("div",xt,[R(s(oe),{value:l.value.username,"onUpdate:value":p[0]||(p[0]=o=>l.value.username=o),placeholder:s(m).auth.username,autofocus:"",onKeyup:re(g,["enter"])},null,8,["value","placeholder"]),R(s(oe),{value:l.value.password,"onUpdate:value":p[1]||(p[1]=o=>l.value.password=o),type:"password","show-password-on":"click",placeholder:s(m).auth.password,onKeyup:re(g,["enter"])},null,8,["value","placeholder"]),R(s(X),{type:"primary",block:"",loading:b.value,onClick:g},{default:T(()=>[L(F(s(m).nav.login),1)]),_:1},8,["loading"]),s(f).query.next?(P(),Q("div",Bt,[L(F(s(m).auth.redirectHint),1),W("code",null,F(s(f).query.next),1)])):he("",!0)]))}}),Et=fe(Rt,[["__scopeId","data-v-1cacc662"]]),_t={class:"header-inner"},Ot={class:"header-actions"},Tt={class:"user-nick text-sm opacity-80"},Ft=U({__name:"AppLayout",setup(e){const t=ue(),r=rt(),u=me(),f=B(()=>u.matched.some(o=>o.meta.wide)),h=O(!1),l=()=>{h.value=!1;const o=typeof u.query.next=="string"?u.query.next:"";o&&o!=="/"&&!o.startsWith("//")&&r.push(o)},b=B(()=>[{label:m.nav.problems,key:"/problems"},{label:m.nav.problemsets,key:"/problemsets"},{label:m.nav.submissions,key:"/submissions"},{label:m.nav.ranking,key:"/ranking"},{label:m.nav.me,key:"/me"}]),g=B(()=>{const o=u.path;return o.startsWith("/problemsets")?"/problemsets":o.startsWith("/problems")?"/problems":o.startsWith("/submissions")?"/submissions":o.startsWith("/ranking")?"/ranking":o.startsWith("/me")?"/me":""}),S=o=>{o!==u.path&&r.push(o)},p=()=>{t.logout(),r.replace("/")};return(o,w)=>(P(),G(s(st),{class:"min-h-screen"},{default:T(()=>[R(s(it),{bordered:"",class:"app-header"},{default:T(()=>{var z,y;return[W("div",_t,[W("div",{class:"text-lg font-bold mr-8 cursor-pointer",onClick:w[0]||(w[0]=M=>s(r).push("/"))},F(s(m).nav.appName),1),R(s(at),{mode:"horizontal",options:b.value,value:g.value,class:"flex-1 app-nav-menu","onUpdate:value":S},null,8,["options","value"]),W("div",Ot,[s(t).isLoggedIn?(P(),Q(Ze,{key:0},[W("span",Tt,F(((z=s(t).user)==null?void 0:z.name)||((y=s(t).user)==null?void 0:y.username)),1),s(t).isAdmin?(P(),G(s(X),{key:0,size:"small",type:"primary",ghost:"",onClick:w[1]||(w[1]=M=>s(r).push("/admin"))},{default:T(()=>[L(F(s(m).nav.adminPanel),1)]),_:1})):he("",!0),R(s(X),{size:"small",onClick:p},{default:T(()=>[L(F(s(m).nav.logout),1)]),_:1})],64)):(P(),G(s(X),{key:1,size:"small",type:"primary",onClick:w[2]||(w[2]=M=>h.value=!0)},{default:T(()=>[L(F(s(m).nav.login),1)]),_:1}))])])]}),_:1}),W("div",{class:tt(["content-wrap",{wide:f.value}])},[R(s(et))],2),R(s($t),{show:h.value,"onUpdate:show":w[3]||(w[3]=z=>h.value=z),width:360,placement:"right"},{default:T(()=>[R(s(kt),{title:s(m).auth.loginTitle,closable:""},{default:T(()=>[R(Et,{onSuccess:l})]),_:1},8,["title"])]),_:1},8,["show"])]),_:1}))}}),Yt=fe(Ft,[["__scopeId","data-v-ef3b1de2"]]);export{Yt as default};
