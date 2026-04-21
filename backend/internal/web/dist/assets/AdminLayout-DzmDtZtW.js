import{t as l,x as m,y as a,q as k,d as P,h as c,ac as X,S as K,r as N,ad as J,u as Q,C as Y,E as Z,f as S,i as G,I as R,J as A,p as ee,M as te,P as oe,Q as h,W as x,R as n,O as re,Z as C,Y as T,V as D,X as F,a2 as se,a4 as ae,$ as le}from"./index-GV3JmupW.js";import{t as u}from"./index-DGOhw6Ge.js";import{c as ne,p as ie,l as de,d as ce,e as ue,b as be,N as ge,a as me}from"./Menu-BDu7m828.js";import{C as he}from"./Dropdown-DNc5_MKS.js";import{f as j,u as pe}from"./get-CjhBgZP8.js";import{_ as fe}from"./_plugin-vue_export-helper-DlAUqK2U.js";import"./Popover-B_1KmcmA.js";import"./use-compitable-BO06rvls.js";const ve=ne(!0),ye=l("layout-sider",`
 flex-shrink: 0;
 box-sizing: border-box;
 position: relative;
 z-index: 1;
 color: var(--n-text-color);
 transition:
 color .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 min-width .3s var(--n-bezier),
 max-width .3s var(--n-bezier),
 transform .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 background-color: var(--n-color);
 display: flex;
 justify-content: flex-end;
`,[m("bordered",[a("border",`
 content: "";
 position: absolute;
 top: 0;
 bottom: 0;
 width: 1px;
 background-color: var(--n-border-color);
 transition: background-color .3s var(--n-bezier);
 `)]),a("left-placement",[m("bordered",[a("border",`
 right: 0;
 `)])]),m("right-placement",`
 justify-content: flex-start;
 `,[m("bordered",[a("border",`
 left: 0;
 `)]),m("collapsed",[l("layout-toggle-button",[l("base-icon",`
 transform: rotate(180deg);
 `)]),l("layout-toggle-bar",[k("&:hover",[a("top",{transform:"rotate(-12deg) scale(1.15) translateY(-2px)"}),a("bottom",{transform:"rotate(12deg) scale(1.15) translateY(2px)"})])])]),l("layout-toggle-button",`
 left: 0;
 transform: translateX(-50%) translateY(-50%);
 `,[l("base-icon",`
 transform: rotate(0);
 `)]),l("layout-toggle-bar",`
 left: -28px;
 transform: rotate(180deg);
 `,[k("&:hover",[a("top",{transform:"rotate(12deg) scale(1.15) translateY(-2px)"}),a("bottom",{transform:"rotate(-12deg) scale(1.15) translateY(2px)"})])])]),m("collapsed",[l("layout-toggle-bar",[k("&:hover",[a("top",{transform:"rotate(-12deg) scale(1.15) translateY(-2px)"}),a("bottom",{transform:"rotate(12deg) scale(1.15) translateY(2px)"})])]),l("layout-toggle-button",[l("base-icon",`
 transform: rotate(0);
 `)])]),l("layout-toggle-button",`
 transition:
 color .3s var(--n-bezier),
 right .3s var(--n-bezier),
 left .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 cursor: pointer;
 width: 24px;
 height: 24px;
 position: absolute;
 top: 50%;
 right: 0;
 border-radius: 50%;
 display: flex;
 align-items: center;
 justify-content: center;
 font-size: 18px;
 color: var(--n-toggle-button-icon-color);
 border: var(--n-toggle-button-border);
 background-color: var(--n-toggle-button-color);
 box-shadow: 0 2px 4px 0px rgba(0, 0, 0, .06);
 transform: translateX(50%) translateY(-50%);
 z-index: 1;
 `,[l("base-icon",`
 transition: transform .3s var(--n-bezier);
 transform: rotate(180deg);
 `)]),l("layout-toggle-bar",`
 cursor: pointer;
 height: 72px;
 width: 32px;
 position: absolute;
 top: calc(50% - 36px);
 right: -28px;
 `,[a("top, bottom",`
 position: absolute;
 width: 4px;
 border-radius: 2px;
 height: 38px;
 left: 14px;
 transition: 
 background-color .3s var(--n-bezier),
 transform .3s var(--n-bezier);
 `),a("bottom",`
 position: absolute;
 top: 34px;
 `),k("&:hover",[a("top",{transform:"rotate(12deg) scale(1.15) translateY(-2px)"}),a("bottom",{transform:"rotate(-12deg) scale(1.15) translateY(2px)"})]),a("top, bottom",{backgroundColor:"var(--n-toggle-bar-color)"}),k("&:hover",[a("top, bottom",{backgroundColor:"var(--n-toggle-bar-color-hover)"})])]),a("border",`
 position: absolute;
 top: 0;
 right: 0;
 bottom: 0;
 width: 1px;
 transition: background-color .3s var(--n-bezier);
 `),l("layout-sider-scroll-container",`
 flex-grow: 1;
 flex-shrink: 0;
 box-sizing: border-box;
 height: 100%;
 opacity: 0;
 transition: opacity .3s var(--n-bezier);
 max-width: 100%;
 `),m("show-content",[l("layout-sider-scroll-container",{opacity:1})]),m("absolute-positioned",`
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 `)]),xe=P({props:{clsPrefix:{type:String,required:!0},onClick:Function},render(){const{clsPrefix:e}=this;return c("div",{onClick:this.onClick,class:`${e}-layout-toggle-bar`},c("div",{class:`${e}-layout-toggle-bar__top`}),c("div",{class:`${e}-layout-toggle-bar__bottom`}))}}),Ce=P({name:"LayoutToggleButton",props:{clsPrefix:{type:String,required:!0},onClick:Function},render(){const{clsPrefix:e}=this;return c("div",{class:`${e}-layout-toggle-button`,onClick:this.onClick},c(X,{clsPrefix:e},{default:()=>c(he,null)}))}}),Se={position:ie,bordered:Boolean,collapsedWidth:{type:Number,default:48},width:{type:[Number,String],default:272},contentClass:String,contentStyle:{type:[String,Object],default:""},collapseMode:{type:String,default:"transform"},collapsed:{type:Boolean,default:void 0},defaultCollapsed:Boolean,showCollapsedContent:{type:Boolean,default:!0},showTrigger:{type:[Boolean,String],default:!1},nativeScrollbar:{type:Boolean,default:!0},inverted:Boolean,scrollbarProps:Object,triggerClass:String,triggerStyle:[String,Object],collapsedTriggerClass:String,collapsedTriggerStyle:[String,Object],"onUpdate:collapsed":[Function,Array],onUpdateCollapsed:[Function,Array],onAfterEnter:Function,onAfterLeave:Function,onExpand:[Function,Array],onCollapse:[Function,Array],onScroll:Function},_e=P({name:"LayoutSider",props:Object.assign(Object.assign({},Y.props),Se),setup(e){const r=G(ce),i=N(null),g=N(null),w=N(e.defaultCollapsed),p=pe(A(e,"collapsed"),w),$=S(()=>j(p.value?e.collapsedWidth:e.width)),I=S(()=>e.collapseMode!=="transform"?{}:{minWidth:j(e.width)}),d=S(()=>r?r.siderPlacement:"left");function f(s,t){if(e.nativeScrollbar){const{value:o}=i;o&&(t===void 0?o.scrollTo(s):o.scrollTo(s,t))}else{const{value:o}=g;o&&o.scrollTo(s,t)}}function z(){const{"onUpdate:collapsed":s,onUpdateCollapsed:t,onExpand:o,onCollapse:B}=e,{value:y}=p;t&&R(t,!y),s&&R(s,!y),w.value=!y,y?o&&R(o):B&&R(B)}let _=0,L=0;const q=s=>{var t;const o=s.target;_=o.scrollLeft,L=o.scrollTop,(t=e.onScroll)===null||t===void 0||t.call(e,s)};J(()=>{if(e.nativeScrollbar){const s=i.value;s&&(s.scrollTop=L,s.scrollLeft=_)}}),ee(ue,{collapsedRef:p,collapseModeRef:A(e,"collapseMode")});const{mergedClsPrefixRef:O,inlineThemeDisabled:W}=Q(e),E=Y("Layout","-layout-sider",ye,de,e,O);function V(s){var t,o;s.propertyName==="max-width"&&(p.value?(t=e.onAfterLeave)===null||t===void 0||t.call(e):(o=e.onAfterEnter)===null||o===void 0||o.call(e))}const U={scrollTo:f},M=S(()=>{const{common:{cubicBezierEaseInOut:s},self:t}=E.value,{siderToggleButtonColor:o,siderToggleButtonBorder:B,siderToggleBarColor:y,siderToggleBarColorHover:H}=t,b={"--n-bezier":s,"--n-toggle-button-color":o,"--n-toggle-button-border":B,"--n-toggle-bar-color":y,"--n-toggle-bar-color-hover":H};return e.inverted?(b["--n-color"]=t.siderColorInverted,b["--n-text-color"]=t.textColorInverted,b["--n-border-color"]=t.siderBorderColorInverted,b["--n-toggle-button-icon-color"]=t.siderToggleButtonIconColorInverted,b.__invertScrollbar=t.__invertScrollbar):(b["--n-color"]=t.siderColor,b["--n-text-color"]=t.textColor,b["--n-border-color"]=t.siderBorderColor,b["--n-toggle-button-icon-color"]=t.siderToggleButtonIconColor),b}),v=W?Z("layout-sider",S(()=>e.inverted?"a":"b"),M,e):void 0;return Object.assign({scrollableElRef:i,scrollbarInstRef:g,mergedClsPrefix:O,mergedTheme:E,styleMaxWidth:$,mergedCollapsed:p,scrollContainerStyle:I,siderPlacement:d,handleNativeElScroll:q,handleTransitionend:V,handleTriggerClick:z,inlineThemeDisabled:W,cssVars:M,themeClass:v==null?void 0:v.themeClass,onRender:v==null?void 0:v.onRender},U)},render(){var e;const{mergedClsPrefix:r,mergedCollapsed:i,showTrigger:g}=this;return(e=this.onRender)===null||e===void 0||e.call(this),c("aside",{class:[`${r}-layout-sider`,this.themeClass,`${r}-layout-sider--${this.position}-positioned`,`${r}-layout-sider--${this.siderPlacement}-placement`,this.bordered&&`${r}-layout-sider--bordered`,i&&`${r}-layout-sider--collapsed`,(!i||this.showCollapsedContent)&&`${r}-layout-sider--show-content`],onTransitionend:this.handleTransitionend,style:[this.inlineThemeDisabled?void 0:this.cssVars,{maxWidth:this.styleMaxWidth,width:j(this.width)}]},this.nativeScrollbar?c("div",{class:[`${r}-layout-sider-scroll-container`,this.contentClass],onScroll:this.handleNativeElScroll,style:[this.scrollContainerStyle,{overflow:"auto"},this.contentStyle],ref:"scrollableElRef"},this.$slots):c(K,Object.assign({},this.scrollbarProps,{onScroll:this.onScroll,ref:"scrollbarInstRef",style:this.scrollContainerStyle,contentStyle:this.contentStyle,contentClass:this.contentClass,theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar,builtinThemeOverrides:this.inverted&&this.cssVars.__invertScrollbar==="true"?{colorHover:"rgba(255, 255, 255, .4)",color:"rgba(255, 255, 255, .3)"}:void 0}),this.$slots),g?g==="bar"?c(xe,{clsPrefix:r,class:i?this.collapsedTriggerClass:this.triggerClass,style:i?this.collapsedTriggerStyle:this.triggerStyle,onClick:this.handleTriggerClick}):c(Ce,{clsPrefix:r,class:i?this.collapsedTriggerClass:this.triggerClass,style:i?this.collapsedTriggerStyle:this.triggerStyle,onClick:this.handleTriggerClick}):null,this.bordered?c("div",{class:`${r}-layout-sider__border`}):null)}}),ke={class:"admin-shell"},Te={class:"header-inner"},we={class:"text-lg font-bold mr-3"},ze={class:"text-sm opacity-70"},Be={class:"header-actions"},Re={class:"user-nick text-sm opacity-80"},Pe=P({__name:"AdminLayout",setup(e){const r=te(),i=ae(),g=le(),w=[{label:u.adminDashboard.overview,key:"/admin"},{label:u.adminDashboard.quickUsers,key:"/admin/users"},{label:u.adminDashboard.quickTags,key:"/admin/tags"},{label:u.adminDashboard.quickProblems,key:"/admin/problems"},{label:u.adminDashboard.quickProblemsets,key:"/admin/problemsets"},{label:u.adminDashboard.quickSubmissions,key:"/admin/submissions"},{label:u.adminDashboard.quickAi,key:"/admin/ai"}],p=S(()=>{const d=g.path;return d.startsWith("/admin/problemsets")?"/admin/problemsets":d.startsWith("/admin/problems")?"/admin/problems":d.startsWith("/admin/tags")?"/admin/tags":d.startsWith("/admin/users")?"/admin/users":d.startsWith("/admin/submissions")?"/admin/submissions":d.startsWith("/admin/ai")?"/admin/ai":"/admin"}),$=d=>{d!==g.path&&i.push(d)},I=()=>{r.logout(),i.replace("/")};return(d,f)=>(re(),oe("div",ke,[h(n(be),{bordered:"",class:"admin-header"},{default:x(()=>{var z,_;return[C("div",Te,[C("div",we,T(n(u).nav.appName),1),C("div",ze,T(n(u).nav.adminConsole),1),f[1]||(f[1]=C("div",{class:"flex-1"},null,-1)),C("div",Be,[C("span",Re,T(((z=n(r).user)==null?void 0:z.name)||((_=n(r).user)==null?void 0:_.username)),1),h(n(D),{size:"small",type:"primary",ghost:"",onClick:f[0]||(f[0]=L=>n(i).push("/problems"))},{default:x(()=>[F(T(n(u).nav.backToFront),1)]),_:1}),h(n(D),{size:"small",onClick:I},{default:x(()=>[F(T(n(u).nav.logout),1)]),_:1})])])]}),_:1}),h(n(ge),{"has-sider":"",class:"admin-body"},{default:x(()=>[h(n(_e),{bordered:"",width:220,"native-scrollbar":!1},{default:x(()=>[h(n(me),{options:w,value:p.value,indent:18,"onUpdate:value":$},null,8,["value"])]),_:1}),h(n(ve),{class:"admin-content","native-scrollbar":!1},{default:x(()=>[h(n(se))]),_:1})]),_:1})]))}}),Me=fe(Pe,[["__scopeId","data-v-9d292489"]]);export{Me as default};
