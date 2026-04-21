import{d as T,h as m,ae as Be,af as Le,ag as Me,ah as ue,ai as _,t as g,x as w,S as $e,u as oe,C as M,ad as ke,E as te,f as b,r as E,p as K,q as C,y as x,aj as k,ak as Ke,al as L,ac as _e,i as B,am as re,an as Z,a1 as je,ao as X,ap as Ve,aq as De,b as ve,ar as Ue,m as Ge,as as qe,I as O,J as me}from"./index-GV3JmupW.js";import{N as We,a as Je}from"./Dropdown-DNc5_MKS.js";import{u as he}from"./get-CjhBgZP8.js";import{V as Xe,c as Y}from"./Popover-B_1KmcmA.js";import{u as Ye}from"./use-compitable-BO06rvls.js";const Ze=T({name:"ChevronDownFilled",render(){return m("svg",{viewBox:"0 0 16 16",fill:"none",xmlns:"http://www.w3.org/2000/svg"},m("path",{d:"M3.20041 5.73966C3.48226 5.43613 3.95681 5.41856 4.26034 5.70041L8 9.22652L11.7397 5.70041C12.0432 5.41856 12.5177 5.43613 12.7996 5.73966C13.0815 6.0432 13.0639 6.51775 12.7603 6.7996L8.51034 10.7996C8.22258 11.0668 7.77743 11.0668 7.48967 10.7996L3.23966 6.7996C2.93613 6.51775 2.91856 6.0432 3.20041 5.73966Z",fill:"currentColor"}))}});function Qe(e){const{baseColor:r,textColor2:t,bodyColor:c,cardColor:l,dividerColor:i,actionColor:s,scrollbarColor:v,scrollbarColorHover:a,invertedColor:f}=e;return{textColor:t,textColorInverted:"#FFF",color:c,colorEmbedded:s,headerColor:l,headerColorInverted:f,footerColor:s,footerColorInverted:f,headerBorderColor:i,headerBorderColorInverted:f,footerBorderColor:i,footerBorderColorInverted:f,siderBorderColor:i,siderBorderColorInverted:f,siderColor:l,siderColorInverted:f,siderToggleButtonBorder:`1px solid ${i}`,siderToggleButtonColor:r,siderToggleButtonIconColor:t,siderToggleButtonIconColorInverted:t,siderToggleBarColor:ue(c,v),siderToggleBarColorHover:ue(c,a),__invertScrollbar:"true"}}const ge=Be({name:"Layout",common:Me,peers:{Scrollbar:Le},self:Qe}),eo=_("n-layout-sider"),xe={type:String,default:"static"},oo=g("layout",`
 color: var(--n-text-color);
 background-color: var(--n-color);
 box-sizing: border-box;
 position: relative;
 z-index: auto;
 flex: auto;
 overflow: hidden;
 transition:
 box-shadow .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
`,[g("layout-scroll-container",`
 overflow-x: hidden;
 box-sizing: border-box;
 height: 100%;
 `),w("absolute-positioned",`
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `)]),to={embedded:Boolean,position:xe,nativeScrollbar:{type:Boolean,default:!0},scrollbarProps:Object,onScroll:Function,contentClass:String,contentStyle:{type:[String,Object],default:""},hasSider:Boolean,siderPlacement:{type:String,default:"left"}},ro=_("n-layout");function no(e){return T({name:e?"LayoutContent":"Layout",props:Object.assign(Object.assign({},M.props),to),setup(r){const t=E(null),c=E(null),{mergedClsPrefixRef:l,inlineThemeDisabled:i}=oe(r),s=M("Layout","-layout",oo,ge,r,l);function v(y,I){if(r.nativeScrollbar){const{value:S}=t;S&&(I===void 0?S.scrollTo(y):S.scrollTo(y,I))}else{const{value:S}=c;S&&S.scrollTo(y,I)}}K(ro,r);let a=0,f=0;const A=y=>{var I;const S=y.target;a=S.scrollLeft,f=S.scrollTop,(I=r.onScroll)===null||I===void 0||I.call(r,y)};ke(()=>{if(r.nativeScrollbar){const y=t.value;y&&(y.scrollTop=f,y.scrollLeft=a)}});const N={display:"flex",flexWrap:"nowrap",width:"100%",flexDirection:"row"},h={scrollTo:v},R=b(()=>{const{common:{cubicBezierEaseInOut:y},self:I}=s.value;return{"--n-bezier":y,"--n-color":r.embedded?I.colorEmbedded:I.color,"--n-text-color":I.textColor}}),P=i?te("layout",b(()=>r.embedded?"e":""),R,r):void 0;return Object.assign({mergedClsPrefix:l,scrollableElRef:t,scrollbarInstRef:c,hasSiderStyle:N,mergedTheme:s,handleNativeElScroll:A,cssVars:i?void 0:R,themeClass:P==null?void 0:P.themeClass,onRender:P==null?void 0:P.onRender},h)},render(){var r;const{mergedClsPrefix:t,hasSider:c}=this;(r=this.onRender)===null||r===void 0||r.call(this);const l=c?this.hasSiderStyle:void 0,i=[this.themeClass,e&&`${t}-layout-content`,`${t}-layout`,`${t}-layout--${this.position}-positioned`];return m("div",{class:i,style:this.cssVars},this.nativeScrollbar?m("div",{ref:"scrollableElRef",class:[`${t}-layout-scroll-container`,this.contentClass],style:[this.contentStyle,l],onScroll:this.handleNativeElScroll},this.$slots):m($e,Object.assign({},this.scrollbarProps,{onScroll:this.onScroll,ref:"scrollbarInstRef",theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar,contentClass:this.contentClass,contentStyle:[this.contentStyle,l]}),this.$slots))}})}const zo=no(!1),io=g("layout-header",`
 transition:
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 box-sizing: border-box;
 width: 100%;
 background-color: var(--n-color);
 color: var(--n-text-color);
`,[w("absolute-positioned",`
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 `),w("bordered",`
 border-bottom: solid 1px var(--n-border-color);
 `)]),lo={position:xe,inverted:Boolean,bordered:{type:Boolean,default:!1}},Io=T({name:"LayoutHeader",props:Object.assign(Object.assign({},M.props),lo),setup(e){const{mergedClsPrefixRef:r,inlineThemeDisabled:t}=oe(e),c=M("Layout","-layout-header",io,ge,e,r),l=b(()=>{const{common:{cubicBezierEaseInOut:s},self:v}=c.value,a={"--n-bezier":s};return e.inverted?(a["--n-color"]=v.headerColorInverted,a["--n-text-color"]=v.textColorInverted,a["--n-border-color"]=v.headerBorderColorInverted):(a["--n-color"]=v.headerColor,a["--n-text-color"]=v.textColor,a["--n-border-color"]=v.headerBorderColor),a}),i=t?te("layout-header",b(()=>e.inverted?"a":"b"),l,e):void 0;return{mergedClsPrefix:r,cssVars:t?void 0:l,themeClass:i==null?void 0:i.themeClass,onRender:i==null?void 0:i.onRender}},render(){var e;const{mergedClsPrefix:r}=this;return(e=this.onRender)===null||e===void 0||e.call(this),m("div",{class:[`${r}-layout-header`,this.themeClass,this.position&&`${r}-layout-header--${this.position}-positioned`,this.bordered&&`${r}-layout-header--bordered`],style:this.cssVars},this.$slots)}}),j=_("n-menu"),be=_("n-submenu"),ne=_("n-menu-item-group"),fe=[C("&::before","background-color: var(--n-item-color-hover);"),x("arrow",`
 color: var(--n-arrow-color-hover);
 `),x("icon",`
 color: var(--n-item-icon-color-hover);
 `),g("menu-item-content-header",`
 color: var(--n-item-text-color-hover);
 `,[C("a",`
 color: var(--n-item-text-color-hover);
 `),x("extra",`
 color: var(--n-item-text-color-hover);
 `)])],pe=[x("icon",`
 color: var(--n-item-icon-color-hover-horizontal);
 `),g("menu-item-content-header",`
 color: var(--n-item-text-color-hover-horizontal);
 `,[C("a",`
 color: var(--n-item-text-color-hover-horizontal);
 `),x("extra",`
 color: var(--n-item-text-color-hover-horizontal);
 `)])],ao=C([g("menu",`
 background-color: var(--n-color);
 color: var(--n-item-text-color);
 overflow: hidden;
 transition: background-color .3s var(--n-bezier);
 box-sizing: border-box;
 font-size: var(--n-font-size);
 padding-bottom: 6px;
 `,[w("horizontal",`
 max-width: 100%;
 width: 100%;
 display: flex;
 overflow: hidden;
 padding-bottom: 0;
 `,[g("submenu","margin: 0;"),g("menu-item","margin: 0;"),g("menu-item-content",`
 padding: 0 20px;
 border-bottom: 2px solid #0000;
 `,[C("&::before","display: none;"),w("selected","border-bottom: 2px solid var(--n-border-color-horizontal)")]),g("menu-item-content",[w("selected",[x("icon","color: var(--n-item-icon-color-active-horizontal);"),g("menu-item-content-header",`
 color: var(--n-item-text-color-active-horizontal);
 `,[C("a","color: var(--n-item-text-color-active-horizontal);"),x("extra","color: var(--n-item-text-color-active-horizontal);")])]),w("child-active",`
 border-bottom: 2px solid var(--n-border-color-horizontal);
 `,[g("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-horizontal);
 `,[C("a",`
 color: var(--n-item-text-color-child-active-horizontal);
 `),x("extra",`
 color: var(--n-item-text-color-child-active-horizontal);
 `)]),x("icon",`
 color: var(--n-item-icon-color-child-active-horizontal);
 `)]),k("disabled",[k("selected, child-active",[C("&:focus-within",pe)]),w("selected",[F(null,[x("icon","color: var(--n-item-icon-color-active-hover-horizontal);"),g("menu-item-content-header",`
 color: var(--n-item-text-color-active-hover-horizontal);
 `,[C("a","color: var(--n-item-text-color-active-hover-horizontal);"),x("extra","color: var(--n-item-text-color-active-hover-horizontal);")])])]),w("child-active",[F(null,[x("icon","color: var(--n-item-icon-color-child-active-hover-horizontal);"),g("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-hover-horizontal);
 `,[C("a","color: var(--n-item-text-color-child-active-hover-horizontal);"),x("extra","color: var(--n-item-text-color-child-active-hover-horizontal);")])])]),F("border-bottom: 2px solid var(--n-border-color-horizontal);",pe)]),g("menu-item-content-header",[C("a","color: var(--n-item-text-color-horizontal);")])])]),k("responsive",[g("menu-item-content-header",`
 overflow: hidden;
 text-overflow: ellipsis;
 `)]),w("collapsed",[g("menu-item-content",[w("selected",[C("&::before",`
 background-color: var(--n-item-color-active-collapsed) !important;
 `)]),g("menu-item-content-header","opacity: 0;"),x("arrow","opacity: 0;"),x("icon","color: var(--n-item-icon-color-collapsed);")])]),g("menu-item",`
 height: var(--n-item-height);
 margin-top: 6px;
 position: relative;
 `),g("menu-item-content",`
 box-sizing: border-box;
 line-height: 1.75;
 height: 100%;
 display: grid;
 grid-template-areas: "icon content arrow";
 grid-template-columns: auto 1fr auto;
 align-items: center;
 cursor: pointer;
 position: relative;
 padding-right: 18px;
 transition:
 background-color .3s var(--n-bezier),
 padding-left .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 `,[C("> *","z-index: 1;"),C("&::before",`
 z-index: auto;
 content: "";
 background-color: #0000;
 position: absolute;
 left: 8px;
 right: 8px;
 top: 0;
 bottom: 0;
 pointer-events: none;
 border-radius: var(--n-border-radius);
 transition: background-color .3s var(--n-bezier);
 `),w("disabled",`
 opacity: .45;
 cursor: not-allowed;
 `),w("collapsed",[x("arrow","transform: rotate(0);")]),w("selected",[C("&::before","background-color: var(--n-item-color-active);"),x("arrow","color: var(--n-arrow-color-active);"),x("icon","color: var(--n-item-icon-color-active);"),g("menu-item-content-header",`
 color: var(--n-item-text-color-active);
 `,[C("a","color: var(--n-item-text-color-active);"),x("extra","color: var(--n-item-text-color-active);")])]),w("child-active",[g("menu-item-content-header",`
 color: var(--n-item-text-color-child-active);
 `,[C("a",`
 color: var(--n-item-text-color-child-active);
 `),x("extra",`
 color: var(--n-item-text-color-child-active);
 `)]),x("arrow",`
 color: var(--n-arrow-color-child-active);
 `),x("icon",`
 color: var(--n-item-icon-color-child-active);
 `)]),k("disabled",[k("selected, child-active",[C("&:focus-within",fe)]),w("selected",[F(null,[x("arrow","color: var(--n-arrow-color-active-hover);"),x("icon","color: var(--n-item-icon-color-active-hover);"),g("menu-item-content-header",`
 color: var(--n-item-text-color-active-hover);
 `,[C("a","color: var(--n-item-text-color-active-hover);"),x("extra","color: var(--n-item-text-color-active-hover);")])])]),w("child-active",[F(null,[x("arrow","color: var(--n-arrow-color-child-active-hover);"),x("icon","color: var(--n-item-icon-color-child-active-hover);"),g("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-hover);
 `,[C("a","color: var(--n-item-text-color-child-active-hover);"),x("extra","color: var(--n-item-text-color-child-active-hover);")])])]),w("selected",[F(null,[C("&::before","background-color: var(--n-item-color-active-hover);")])]),F(null,fe)]),x("icon",`
 grid-area: icon;
 color: var(--n-item-icon-color);
 transition:
 color .3s var(--n-bezier),
 font-size .3s var(--n-bezier),
 margin-right .3s var(--n-bezier);
 box-sizing: content-box;
 display: inline-flex;
 align-items: center;
 justify-content: center;
 `),x("arrow",`
 grid-area: arrow;
 font-size: 16px;
 color: var(--n-arrow-color);
 transform: rotate(180deg);
 opacity: 1;
 transition:
 color .3s var(--n-bezier),
 transform 0.2s var(--n-bezier),
 opacity 0.2s var(--n-bezier);
 `),g("menu-item-content-header",`
 grid-area: content;
 transition:
 color .3s var(--n-bezier),
 opacity .3s var(--n-bezier);
 opacity: 1;
 white-space: nowrap;
 color: var(--n-item-text-color);
 `,[C("a",`
 outline: none;
 text-decoration: none;
 transition: color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 `,[C("&::before",`
 content: "";
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `)]),x("extra",`
 font-size: .93em;
 color: var(--n-group-text-color);
 transition: color .3s var(--n-bezier);
 `)])]),g("submenu",`
 cursor: pointer;
 position: relative;
 margin-top: 6px;
 `,[g("menu-item-content",`
 height: var(--n-item-height);
 `),g("submenu-children",`
 overflow: hidden;
 padding: 0;
 `,[Ke({duration:".2s"})])]),g("menu-item-group",[g("menu-item-group-title",`
 margin-top: 6px;
 color: var(--n-group-text-color);
 cursor: default;
 font-size: .93em;
 height: 36px;
 display: flex;
 align-items: center;
 transition:
 padding-left .3s var(--n-bezier),
 color .3s var(--n-bezier);
 `)])]),g("menu-tooltip",[C("a",`
 color: inherit;
 text-decoration: none;
 `)]),g("menu-divider",`
 transition: background-color .3s var(--n-bezier);
 background-color: var(--n-divider-color);
 height: 1px;
 margin: 6px 18px;
 `)]);function F(e,r){return[w("hover",e,r),C("&:hover",e,r)]}const Ce=T({name:"MenuOptionContent",props:{collapsed:Boolean,disabled:Boolean,title:[String,Function],icon:Function,extra:[String,Function],showArrow:Boolean,childActive:Boolean,hover:Boolean,paddingLeft:Number,selected:Boolean,maxIconSize:{type:Number,required:!0},activeIconSize:{type:Number,required:!0},iconMarginRight:{type:Number,required:!0},clsPrefix:{type:String,required:!0},onClick:Function,tmNode:{type:Object,required:!0},isEllipsisPlaceholder:Boolean},setup(e){const{props:r}=B(j);return{menuProps:r,style:b(()=>{const{paddingLeft:t}=e;return{paddingLeft:t&&`${t}px`}}),iconStyle:b(()=>{const{maxIconSize:t,activeIconSize:c,iconMarginRight:l}=e;return{width:`${t}px`,height:`${t}px`,fontSize:`${c}px`,marginRight:`${l}px`}})}},render(){const{clsPrefix:e,tmNode:r,menuProps:{renderIcon:t,renderLabel:c,renderExtra:l,expandIcon:i}}=this,s=t?t(r.rawNode):L(this.icon);return m("div",{onClick:v=>{var a;(a=this.onClick)===null||a===void 0||a.call(this,v)},role:"none",class:[`${e}-menu-item-content`,{[`${e}-menu-item-content--selected`]:this.selected,[`${e}-menu-item-content--collapsed`]:this.collapsed,[`${e}-menu-item-content--child-active`]:this.childActive,[`${e}-menu-item-content--disabled`]:this.disabled,[`${e}-menu-item-content--hover`]:this.hover}],style:this.style},s&&m("div",{class:`${e}-menu-item-content__icon`,style:this.iconStyle,role:"none"},[s]),m("div",{class:`${e}-menu-item-content-header`,role:"none"},this.isEllipsisPlaceholder?this.title:c?c(r.rawNode):L(this.title),this.extra||l?m("span",{class:`${e}-menu-item-content-header__extra`}," ",l?l(r.rawNode):L(this.extra)):null),this.showArrow?m(_e,{ariaHidden:!0,class:`${e}-menu-item-content__arrow`,clsPrefix:e},{default:()=>i?i(r.rawNode):m(Ze,null)}):null)}}),G=8;function ie(e){const r=B(j),{props:t,mergedCollapsedRef:c}=r,l=B(be,null),i=B(ne,null),s=b(()=>t.mode==="horizontal"),v=b(()=>s.value?t.dropdownPlacement:"tmNodes"in e?"right-start":"right"),a=b(()=>{var h;return Math.max((h=t.collapsedIconSize)!==null&&h!==void 0?h:t.iconSize,t.iconSize)}),f=b(()=>{var h;return!s.value&&e.root&&c.value&&(h=t.collapsedIconSize)!==null&&h!==void 0?h:t.iconSize}),A=b(()=>{if(s.value)return;const{collapsedWidth:h,indent:R,rootIndent:P}=t,{root:y,isGroup:I}=e,S=P===void 0?R:P;return y?c.value?h/2-a.value/2:S:i&&typeof i.paddingLeftRef.value=="number"?R/2+i.paddingLeftRef.value:l&&typeof l.paddingLeftRef.value=="number"?(I?R/2:R)+l.paddingLeftRef.value:0}),N=b(()=>{const{collapsedWidth:h,indent:R,rootIndent:P}=t,{value:y}=a,{root:I}=e;return s.value||!I||!c.value?G:(P===void 0?R:P)+y+G-(h+y)/2});return{dropdownPlacement:v,activeIconSize:f,maxIconSize:a,paddingLeft:A,iconMarginRight:N,NMenu:r,NSubmenu:l,NMenuOptionGroup:i}}const le={internalKey:{type:[String,Number],required:!0},root:Boolean,isGroup:Boolean,level:{type:Number,required:!0},title:[String,Function],extra:[String,Function]},co=T({name:"MenuDivider",setup(){const e=B(j),{mergedClsPrefixRef:r,isHorizontalRef:t}=e;return()=>t.value?null:m("div",{class:`${r.value}-menu-divider`})}}),ye=Object.assign(Object.assign({},le),{tmNode:{type:Object,required:!0},disabled:Boolean,icon:Function,onClick:Function}),so=re(ye),uo=T({name:"MenuOption",props:ye,setup(e){const r=ie(e),{NSubmenu:t,NMenu:c,NMenuOptionGroup:l}=r,{props:i,mergedClsPrefixRef:s,mergedCollapsedRef:v}=c,a=t?t.mergedDisabledRef:l?l.mergedDisabledRef:{value:!1},f=b(()=>a.value||e.disabled);function A(h){const{onClick:R}=e;R&&R(h)}function N(h){f.value||(c.doSelect(e.internalKey,e.tmNode.rawNode),A(h))}return{mergedClsPrefix:s,dropdownPlacement:r.dropdownPlacement,paddingLeft:r.paddingLeft,iconMarginRight:r.iconMarginRight,maxIconSize:r.maxIconSize,activeIconSize:r.activeIconSize,mergedTheme:c.mergedThemeRef,menuProps:i,dropdownEnabled:Z(()=>e.root&&v.value&&i.mode!=="horizontal"&&!f.value),selected:Z(()=>c.mergedValueRef.value===e.internalKey),mergedDisabled:f,handleClick:N}},render(){const{mergedClsPrefix:e,mergedTheme:r,tmNode:t,menuProps:{renderLabel:c,nodeProps:l}}=this,i=l==null?void 0:l(t.rawNode);return m("div",Object.assign({},i,{role:"menuitem",class:[`${e}-menu-item`,i==null?void 0:i.class]}),m(We,{theme:r.peers.Tooltip,themeOverrides:r.peerOverrides.Tooltip,trigger:"hover",placement:this.dropdownPlacement,disabled:!this.dropdownEnabled||this.title===void 0,internalExtraClass:["menu-tooltip"]},{default:()=>c?c(t.rawNode):L(this.title),trigger:()=>m(Ce,{tmNode:t,clsPrefix:e,paddingLeft:this.paddingLeft,iconMarginRight:this.iconMarginRight,maxIconSize:this.maxIconSize,activeIconSize:this.activeIconSize,selected:this.selected,title:this.title,extra:this.extra,disabled:this.mergedDisabled,icon:this.icon,onClick:this.handleClick})}))}}),ze=Object.assign(Object.assign({},le),{tmNode:{type:Object,required:!0},tmNodes:{type:Array,required:!0}}),vo=re(ze),mo=T({name:"MenuOptionGroup",props:ze,setup(e){const r=ie(e),{NSubmenu:t}=r,c=b(()=>t!=null&&t.mergedDisabledRef.value?!0:e.tmNode.disabled);K(ne,{paddingLeftRef:r.paddingLeft,mergedDisabledRef:c});const{mergedClsPrefixRef:l,props:i}=B(j);return function(){const{value:s}=l,v=r.paddingLeft.value,{nodeProps:a}=i,f=a==null?void 0:a(e.tmNode.rawNode);return m("div",{class:`${s}-menu-item-group`,role:"group"},m("div",Object.assign({},f,{class:[`${s}-menu-item-group-title`,f==null?void 0:f.class],style:[(f==null?void 0:f.style)||"",v!==void 0?`padding-left: ${v}px;`:""]}),L(e.title),e.extra?m(je,null," ",L(e.extra)):null),m("div",null,e.tmNodes.map(A=>ae(A,i))))}}});function Q(e){return e.type==="divider"||e.type==="render"}function ho(e){return e.type==="divider"}function ae(e,r){const{rawNode:t}=e,{show:c}=t;if(c===!1)return null;if(Q(t))return ho(t)?m(co,Object.assign({key:e.key},t.props)):null;const{labelField:l}=r,{key:i,level:s,isGroup:v}=e,a=Object.assign(Object.assign({},t),{title:t.title||t[l],extra:t.titleExtra||t.extra,key:i,internalKey:i,level:s,root:s===0,isGroup:v});return e.children?e.isGroup?m(mo,X(a,vo,{tmNode:e,tmNodes:e.children,key:i})):m(ee,X(a,fo,{key:i,rawNodes:t[r.childrenField],tmNodes:e.children,tmNode:e})):m(uo,X(a,so,{key:i,tmNode:e}))}const Ie=Object.assign(Object.assign({},le),{rawNodes:{type:Array,default:()=>[]},tmNodes:{type:Array,default:()=>[]},tmNode:{type:Object,required:!0},disabled:Boolean,icon:Function,onClick:Function,domId:String,virtualChildActive:{type:Boolean,default:void 0},isEllipsisPlaceholder:Boolean}),fo=re(Ie),ee=T({name:"Submenu",props:Ie,setup(e){const r=ie(e),{NMenu:t,NSubmenu:c}=r,{props:l,mergedCollapsedRef:i,mergedThemeRef:s}=t,v=b(()=>{const{disabled:h}=e;return c!=null&&c.mergedDisabledRef.value||l.disabled?!0:h}),a=E(!1);K(be,{paddingLeftRef:r.paddingLeft,mergedDisabledRef:v}),K(ne,null);function f(){const{onClick:h}=e;h&&h()}function A(){v.value||(i.value||t.toggleExpand(e.internalKey),f())}function N(h){a.value=h}return{menuProps:l,mergedTheme:s,doSelect:t.doSelect,inverted:t.invertedRef,isHorizontal:t.isHorizontalRef,mergedClsPrefix:t.mergedClsPrefixRef,maxIconSize:r.maxIconSize,activeIconSize:r.activeIconSize,iconMarginRight:r.iconMarginRight,dropdownPlacement:r.dropdownPlacement,dropdownShow:a,paddingLeft:r.paddingLeft,mergedDisabled:v,mergedValue:t.mergedValueRef,childActive:Z(()=>{var h;return(h=e.virtualChildActive)!==null&&h!==void 0?h:t.activePathRef.value.includes(e.internalKey)}),collapsed:b(()=>l.mode==="horizontal"?!1:i.value?!0:!t.mergedExpandedKeysRef.value.includes(e.internalKey)),dropdownEnabled:b(()=>!v.value&&(l.mode==="horizontal"||i.value)),handlePopoverShowChange:N,handleClick:A}},render(){var e;const{mergedClsPrefix:r,menuProps:{renderIcon:t,renderLabel:c}}=this,l=()=>{const{isHorizontal:s,paddingLeft:v,collapsed:a,mergedDisabled:f,maxIconSize:A,activeIconSize:N,title:h,childActive:R,icon:P,handleClick:y,menuProps:{nodeProps:I},dropdownShow:S,iconMarginRight:q,tmNode:$,mergedClsPrefix:V,isEllipsisPlaceholder:W,extra:D}=this,H=I==null?void 0:I($.rawNode);return m("div",Object.assign({},H,{class:[`${V}-menu-item`,H==null?void 0:H.class],role:"menuitem"}),m(Ce,{tmNode:$,paddingLeft:v,collapsed:a,disabled:f,iconMarginRight:q,maxIconSize:A,activeIconSize:N,title:h,extra:D,showArrow:!s,childActive:R,clsPrefix:V,icon:P,hover:S,onClick:y,isEllipsisPlaceholder:W}))},i=()=>m(Ve,null,{default:()=>{const{tmNodes:s,collapsed:v}=this;return v?null:m("div",{class:`${r}-submenu-children`,role:"menu"},s.map(a=>ae(a,this.menuProps)))}});return this.root?m(Je,Object.assign({size:"large",trigger:"hover"},(e=this.menuProps)===null||e===void 0?void 0:e.dropdownProps,{themeOverrides:this.mergedTheme.peerOverrides.Dropdown,theme:this.mergedTheme.peers.Dropdown,builtinThemeOverrides:{fontSizeLarge:"14px",optionIconSizeLarge:"18px"},value:this.mergedValue,disabled:!this.dropdownEnabled,placement:this.dropdownPlacement,keyField:this.menuProps.keyField,labelField:this.menuProps.labelField,childrenField:this.menuProps.childrenField,onUpdateShow:this.handlePopoverShowChange,options:this.rawNodes,onSelect:this.doSelect,inverted:this.inverted,renderIcon:t,renderLabel:c}),{default:()=>m("div",{class:`${r}-submenu`,role:"menu","aria-expanded":!this.collapsed,id:this.domId},l(),this.isHorizontal?null:i())}):m("div",{class:`${r}-submenu`,role:"menu","aria-expanded":!this.collapsed,id:this.domId},l(),i())}}),po=Object.assign(Object.assign({},M.props),{options:{type:Array,default:()=>[]},collapsed:{type:Boolean,default:void 0},collapsedWidth:{type:Number,default:48},iconSize:{type:Number,default:20},collapsedIconSize:{type:Number,default:24},rootIndent:Number,indent:{type:Number,default:32},labelField:{type:String,default:"label"},keyField:{type:String,default:"key"},childrenField:{type:String,default:"children"},disabledField:{type:String,default:"disabled"},defaultExpandAll:Boolean,defaultExpandedKeys:Array,expandedKeys:Array,value:[String,Number],defaultValue:{type:[String,Number],default:null},mode:{type:String,default:"vertical"},watchProps:{type:Array,default:void 0},disabled:Boolean,show:{type:Boolean,default:!0},inverted:Boolean,"onUpdate:expandedKeys":[Function,Array],onUpdateExpandedKeys:[Function,Array],onUpdateValue:[Function,Array],"onUpdate:value":[Function,Array],expandIcon:Function,renderIcon:Function,renderLabel:Function,renderExtra:Function,dropdownProps:Object,accordion:Boolean,nodeProps:Function,dropdownPlacement:{type:String,default:"bottom"},responsive:Boolean,items:Array,onOpenNamesChange:[Function,Array],onSelect:[Function,Array],onExpandedNamesChange:[Function,Array],expandedNames:Array,defaultExpandedNames:Array}),wo=T({name:"Menu",inheritAttrs:!1,props:po,setup(e){const{mergedClsPrefixRef:r,inlineThemeDisabled:t}=oe(e),c=M("Menu","-menu",ao,qe,e,r),l=B(eo,null),i=b(()=>{var d;const{collapsed:p}=e;if(p!==void 0)return p;if(l){const{collapseModeRef:o,collapsedRef:u}=l;if(o.value==="width")return(d=u.value)!==null&&d!==void 0?d:!1}return!1}),s=b(()=>{const{keyField:d,childrenField:p,disabledField:o}=e;return Y(e.items||e.options,{getIgnored(u){return Q(u)},getChildren(u){return u[p]},getDisabled(u){return u[o]},getKey(u){var z;return(z=u[d])!==null&&z!==void 0?z:u.name}})}),v=b(()=>new Set(s.value.treeNodes.map(d=>d.key))),{watchProps:a}=e,f=E(null);a!=null&&a.includes("defaultValue")?ve(()=>{f.value=e.defaultValue}):f.value=e.defaultValue;const A=me(e,"value"),N=he(A,f),h=E([]),R=()=>{h.value=e.defaultExpandAll?s.value.getNonLeafKeys():e.defaultExpandedNames||e.defaultExpandedKeys||s.value.getPath(N.value,{includeSelf:!1}).keyPath};a!=null&&a.includes("defaultExpandedKeys")?ve(R):R();const P=Ye(e,["expandedNames","expandedKeys"]),y=he(P,h),I=b(()=>s.value.treeNodes),S=b(()=>s.value.getPath(N.value).keyPath);K(j,{props:e,mergedCollapsedRef:i,mergedThemeRef:c,mergedValueRef:N,mergedExpandedKeysRef:y,activePathRef:S,mergedClsPrefixRef:r,isHorizontalRef:b(()=>e.mode==="horizontal"),invertedRef:me(e,"inverted"),doSelect:q,toggleExpand:V});function q(d,p){const{"onUpdate:value":o,onUpdateValue:u,onSelect:z}=e;u&&O(u,d,p),o&&O(o,d,p),z&&O(z,d,p),f.value=d}function $(d){const{"onUpdate:expandedKeys":p,onUpdateExpandedKeys:o,onExpandedNamesChange:u,onOpenNamesChange:z}=e;p&&O(p,d),o&&O(o,d),u&&O(u,d),z&&O(z,d),h.value=d}function V(d){const p=Array.from(y.value),o=p.findIndex(u=>u===d);if(~o)p.splice(o,1);else{if(e.accordion&&v.value.has(d)){const u=p.findIndex(z=>v.value.has(z));u>-1&&p.splice(u,1)}p.push(d)}$(p)}const W=d=>{const p=s.value.getPath(d??N.value,{includeSelf:!1}).keyPath;if(!p.length)return;const o=Array.from(y.value),u=new Set([...o,...p]);e.accordion&&v.value.forEach(z=>{u.has(z)&&!p.includes(z)&&u.delete(z)}),$(Array.from(u))},D=b(()=>{const{inverted:d}=e,{common:{cubicBezierEaseInOut:p},self:o}=c.value,{borderRadius:u,borderColorHorizontal:z,fontSize:Ee,itemHeight:Oe,dividerColor:Fe}=o,n={"--n-divider-color":Fe,"--n-bezier":p,"--n-font-size":Ee,"--n-border-color-horizontal":z,"--n-border-radius":u,"--n-item-height":Oe};return d?(n["--n-group-text-color"]=o.groupTextColorInverted,n["--n-color"]=o.colorInverted,n["--n-item-text-color"]=o.itemTextColorInverted,n["--n-item-text-color-hover"]=o.itemTextColorHoverInverted,n["--n-item-text-color-active"]=o.itemTextColorActiveInverted,n["--n-item-text-color-child-active"]=o.itemTextColorChildActiveInverted,n["--n-item-text-color-child-active-hover"]=o.itemTextColorChildActiveInverted,n["--n-item-text-color-active-hover"]=o.itemTextColorActiveHoverInverted,n["--n-item-icon-color"]=o.itemIconColorInverted,n["--n-item-icon-color-hover"]=o.itemIconColorHoverInverted,n["--n-item-icon-color-active"]=o.itemIconColorActiveInverted,n["--n-item-icon-color-active-hover"]=o.itemIconColorActiveHoverInverted,n["--n-item-icon-color-child-active"]=o.itemIconColorChildActiveInverted,n["--n-item-icon-color-child-active-hover"]=o.itemIconColorChildActiveHoverInverted,n["--n-item-icon-color-collapsed"]=o.itemIconColorCollapsedInverted,n["--n-item-text-color-horizontal"]=o.itemTextColorHorizontalInverted,n["--n-item-text-color-hover-horizontal"]=o.itemTextColorHoverHorizontalInverted,n["--n-item-text-color-active-horizontal"]=o.itemTextColorActiveHorizontalInverted,n["--n-item-text-color-child-active-horizontal"]=o.itemTextColorChildActiveHorizontalInverted,n["--n-item-text-color-child-active-hover-horizontal"]=o.itemTextColorChildActiveHoverHorizontalInverted,n["--n-item-text-color-active-hover-horizontal"]=o.itemTextColorActiveHoverHorizontalInverted,n["--n-item-icon-color-horizontal"]=o.itemIconColorHorizontalInverted,n["--n-item-icon-color-hover-horizontal"]=o.itemIconColorHoverHorizontalInverted,n["--n-item-icon-color-active-horizontal"]=o.itemIconColorActiveHorizontalInverted,n["--n-item-icon-color-active-hover-horizontal"]=o.itemIconColorActiveHoverHorizontalInverted,n["--n-item-icon-color-child-active-horizontal"]=o.itemIconColorChildActiveHorizontalInverted,n["--n-item-icon-color-child-active-hover-horizontal"]=o.itemIconColorChildActiveHoverHorizontalInverted,n["--n-arrow-color"]=o.arrowColorInverted,n["--n-arrow-color-hover"]=o.arrowColorHoverInverted,n["--n-arrow-color-active"]=o.arrowColorActiveInverted,n["--n-arrow-color-active-hover"]=o.arrowColorActiveHoverInverted,n["--n-arrow-color-child-active"]=o.arrowColorChildActiveInverted,n["--n-arrow-color-child-active-hover"]=o.arrowColorChildActiveHoverInverted,n["--n-item-color-hover"]=o.itemColorHoverInverted,n["--n-item-color-active"]=o.itemColorActiveInverted,n["--n-item-color-active-hover"]=o.itemColorActiveHoverInverted,n["--n-item-color-active-collapsed"]=o.itemColorActiveCollapsedInverted):(n["--n-group-text-color"]=o.groupTextColor,n["--n-color"]=o.color,n["--n-item-text-color"]=o.itemTextColor,n["--n-item-text-color-hover"]=o.itemTextColorHover,n["--n-item-text-color-active"]=o.itemTextColorActive,n["--n-item-text-color-child-active"]=o.itemTextColorChildActive,n["--n-item-text-color-child-active-hover"]=o.itemTextColorChildActiveHover,n["--n-item-text-color-active-hover"]=o.itemTextColorActiveHover,n["--n-item-icon-color"]=o.itemIconColor,n["--n-item-icon-color-hover"]=o.itemIconColorHover,n["--n-item-icon-color-active"]=o.itemIconColorActive,n["--n-item-icon-color-active-hover"]=o.itemIconColorActiveHover,n["--n-item-icon-color-child-active"]=o.itemIconColorChildActive,n["--n-item-icon-color-child-active-hover"]=o.itemIconColorChildActiveHover,n["--n-item-icon-color-collapsed"]=o.itemIconColorCollapsed,n["--n-item-text-color-horizontal"]=o.itemTextColorHorizontal,n["--n-item-text-color-hover-horizontal"]=o.itemTextColorHoverHorizontal,n["--n-item-text-color-active-horizontal"]=o.itemTextColorActiveHorizontal,n["--n-item-text-color-child-active-horizontal"]=o.itemTextColorChildActiveHorizontal,n["--n-item-text-color-child-active-hover-horizontal"]=o.itemTextColorChildActiveHoverHorizontal,n["--n-item-text-color-active-hover-horizontal"]=o.itemTextColorActiveHoverHorizontal,n["--n-item-icon-color-horizontal"]=o.itemIconColorHorizontal,n["--n-item-icon-color-hover-horizontal"]=o.itemIconColorHoverHorizontal,n["--n-item-icon-color-active-horizontal"]=o.itemIconColorActiveHorizontal,n["--n-item-icon-color-active-hover-horizontal"]=o.itemIconColorActiveHoverHorizontal,n["--n-item-icon-color-child-active-horizontal"]=o.itemIconColorChildActiveHorizontal,n["--n-item-icon-color-child-active-hover-horizontal"]=o.itemIconColorChildActiveHoverHorizontal,n["--n-arrow-color"]=o.arrowColor,n["--n-arrow-color-hover"]=o.arrowColorHover,n["--n-arrow-color-active"]=o.arrowColorActive,n["--n-arrow-color-active-hover"]=o.arrowColorActiveHover,n["--n-arrow-color-child-active"]=o.arrowColorChildActive,n["--n-arrow-color-child-active-hover"]=o.arrowColorChildActiveHover,n["--n-item-color-hover"]=o.itemColorHover,n["--n-item-color-active"]=o.itemColorActive,n["--n-item-color-active-hover"]=o.itemColorActiveHover,n["--n-item-color-active-collapsed"]=o.itemColorActiveCollapsed),n}),H=t?te("menu",b(()=>e.inverted?"a":"b"),D,e):void 0,J=Ue(),ce=E(null),we=E(null);let de=!0;const se=()=>{var d;de?de=!1:(d=ce.value)===null||d===void 0||d.sync({showAllItemsBeforeCalculate:!0})};function Se(){return document.getElementById(J)}const U=E(-1);function Re(d){U.value=e.options.length-d}function Pe(d){d||(U.value=-1)}const Ne=b(()=>{const d=U.value;return{children:d===-1?[]:e.options.slice(d)}}),Ae=b(()=>{const{childrenField:d,disabledField:p,keyField:o}=e;return Y([Ne.value],{getIgnored(u){return Q(u)},getChildren(u){return u[d]},getDisabled(u){return u[p]},getKey(u){var z;return(z=u[o])!==null&&z!==void 0?z:u.name}})}),He=b(()=>Y([{}]).treeNodes[0]);function Te(){var d;if(U.value===-1)return m(ee,{root:!0,level:0,key:"__ellpisisGroupPlaceholder__",internalKey:"__ellpisisGroupPlaceholder__",title:"···",tmNode:He.value,domId:J,isEllipsisPlaceholder:!0});const p=Ae.value.treeNodes[0],o=S.value,u=!!(!((d=p.children)===null||d===void 0)&&d.some(z=>o.includes(z.key)));return m(ee,{level:0,root:!0,key:"__ellpisisGroup__",internalKey:"__ellpisisGroup__",title:"···",virtualChildActive:u,tmNode:p,domId:J,rawNodes:p.rawNode.children||[],tmNodes:p.children||[],isEllipsisPlaceholder:!0})}return{mergedClsPrefix:r,controlledExpandedKeys:P,uncontrolledExpanededKeys:h,mergedExpandedKeys:y,uncontrolledValue:f,mergedValue:N,activePath:S,tmNodes:I,mergedTheme:c,mergedCollapsed:i,cssVars:t?void 0:D,themeClass:H==null?void 0:H.themeClass,overflowRef:ce,counterRef:we,updateCounter:()=>{},onResize:se,onUpdateOverflow:Pe,onUpdateCount:Re,renderCounter:Te,getCounter:Se,onRender:H==null?void 0:H.onRender,showOption:W,deriveResponsiveState:se}},render(){const{mergedClsPrefix:e,mode:r,themeClass:t,onRender:c}=this;c==null||c();const l=()=>this.tmNodes.map(a=>ae(a,this.$props)),s=r==="horizontal"&&this.responsive,v=()=>m("div",Ge(this.$attrs,{role:r==="horizontal"?"menubar":"menu",class:[`${e}-menu`,t,`${e}-menu--${r}`,s&&`${e}-menu--responsive`,this.mergedCollapsed&&`${e}-menu--collapsed`],style:this.cssVars}),s?m(Xe,{ref:"overflowRef",onUpdateOverflow:this.onUpdateOverflow,getCounter:this.getCounter,onUpdateCount:this.onUpdateCount,updateCounter:this.updateCounter,style:{width:"100%",display:"flex",overflow:"hidden"}},{default:l,counter:this.renderCounter}):l());return s?m(De,{onResize:this.onResize},{default:v}):v()}});export{zo as N,wo as a,Io as b,no as c,ro as d,eo as e,ge as l,xe as p};
