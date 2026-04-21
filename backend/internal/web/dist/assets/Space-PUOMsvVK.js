import{ag as ge,cd as ue,b4 as n,t as pe,x as y,y as I,aj as w,q as H,d as A,aF as U,h as z,N as ve,u as K,C as M,a as J,E as be,r as Ce,I as fe,f as $,az as p,ba as me,ce as L,p as ke,J as xe,ai as ye,cf as ze,aD as Se,aw as Pe,cg as Ie,ch as Be,bF as V}from"./index-GV3JmupW.js";function $e(e,c="default",a=[]){const l=e.$slots[c];return l===void 0?a:l()}function Re(e){const{textColor2:c,primaryColorHover:a,primaryColorPressed:u,primaryColor:l,infoColor:h,successColor:i,warningColor:o,errorColor:s,baseColor:b,borderColor:C,opacityDisabled:g,tagColor:B,closeIconColor:P,closeIconColorHover:v,closeIconColorPressed:r,borderRadiusSmall:t,fontSizeMini:f,fontSizeTiny:d,fontSizeSmall:k,fontSizeMedium:x,heightMini:S,heightTiny:m,heightSmall:E,heightMedium:T,closeColorHover:j,closeColorPressed:_,buttonColor2Hover:O,buttonColor2Pressed:W,fontWeightStrong:F}=e;return Object.assign(Object.assign({},ue),{closeBorderRadius:t,heightTiny:S,heightSmall:m,heightMedium:E,heightLarge:T,borderRadius:t,opacityDisabled:g,fontSizeTiny:f,fontSizeSmall:d,fontSizeMedium:k,fontSizeLarge:x,fontWeightStrong:F,textColorCheckable:c,textColorHoverCheckable:c,textColorPressedCheckable:c,textColorChecked:b,colorCheckable:"#0000",colorHoverCheckable:O,colorPressedCheckable:W,colorChecked:l,colorCheckedHover:a,colorCheckedPressed:u,border:`1px solid ${C}`,textColor:c,color:B,colorBordered:"rgb(250, 250, 252)",closeIconColor:P,closeIconColorHover:v,closeIconColorPressed:r,closeColorHover:j,closeColorPressed:_,borderPrimary:`1px solid ${n(l,{alpha:.3})}`,textColorPrimary:l,colorPrimary:n(l,{alpha:.12}),colorBorderedPrimary:n(l,{alpha:.1}),closeIconColorPrimary:l,closeIconColorHoverPrimary:l,closeIconColorPressedPrimary:l,closeColorHoverPrimary:n(l,{alpha:.12}),closeColorPressedPrimary:n(l,{alpha:.18}),borderInfo:`1px solid ${n(h,{alpha:.3})}`,textColorInfo:h,colorInfo:n(h,{alpha:.12}),colorBorderedInfo:n(h,{alpha:.1}),closeIconColorInfo:h,closeIconColorHoverInfo:h,closeIconColorPressedInfo:h,closeColorHoverInfo:n(h,{alpha:.12}),closeColorPressedInfo:n(h,{alpha:.18}),borderSuccess:`1px solid ${n(i,{alpha:.3})}`,textColorSuccess:i,colorSuccess:n(i,{alpha:.12}),colorBorderedSuccess:n(i,{alpha:.1}),closeIconColorSuccess:i,closeIconColorHoverSuccess:i,closeIconColorPressedSuccess:i,closeColorHoverSuccess:n(i,{alpha:.12}),closeColorPressedSuccess:n(i,{alpha:.18}),borderWarning:`1px solid ${n(o,{alpha:.35})}`,textColorWarning:o,colorWarning:n(o,{alpha:.15}),colorBorderedWarning:n(o,{alpha:.12}),closeIconColorWarning:o,closeIconColorHoverWarning:o,closeIconColorPressedWarning:o,closeColorHoverWarning:n(o,{alpha:.12}),closeColorPressedWarning:n(o,{alpha:.18}),borderError:`1px solid ${n(s,{alpha:.23})}`,textColorError:s,colorError:n(s,{alpha:.1}),colorBorderedError:n(s,{alpha:.08}),closeIconColorError:s,closeIconColorHoverError:s,closeIconColorPressedError:s,closeColorHoverError:n(s,{alpha:.12}),closeColorPressedError:n(s,{alpha:.18})})}const we={common:ge,self:Re},He={color:Object,type:{type:String,default:"default"},round:Boolean,size:String,closable:Boolean,disabled:{type:Boolean,default:void 0}},Me=pe("tag",`
 --n-close-margin: var(--n-close-margin-top) var(--n-close-margin-right) var(--n-close-margin-bottom) var(--n-close-margin-left);
 white-space: nowrap;
 position: relative;
 box-sizing: border-box;
 cursor: default;
 display: inline-flex;
 align-items: center;
 flex-wrap: nowrap;
 padding: var(--n-padding);
 border-radius: var(--n-border-radius);
 color: var(--n-text-color);
 background-color: var(--n-color);
 transition: 
 border-color .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier),
 opacity .3s var(--n-bezier);
 line-height: 1;
 height: var(--n-height);
 font-size: var(--n-font-size);
`,[y("strong",`
 font-weight: var(--n-font-weight-strong);
 `),I("border",`
 pointer-events: none;
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 border-radius: inherit;
 border: var(--n-border);
 transition: border-color .3s var(--n-bezier);
 `),I("icon",`
 display: flex;
 margin: 0 4px 0 0;
 color: var(--n-text-color);
 transition: color .3s var(--n-bezier);
 font-size: var(--n-avatar-size-override);
 `),I("avatar",`
 display: flex;
 margin: 0 6px 0 0;
 `),I("close",`
 margin: var(--n-close-margin);
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 `),y("round",`
 padding: 0 calc(var(--n-height) / 3);
 border-radius: calc(var(--n-height) / 2);
 `,[I("icon",`
 margin: 0 4px 0 calc((var(--n-height) - 8px) / -2);
 `),I("avatar",`
 margin: 0 6px 0 calc((var(--n-height) - 8px) / -2);
 `),y("closable",`
 padding: 0 calc(var(--n-height) / 4) 0 calc(var(--n-height) / 3);
 `)]),y("icon, avatar",[y("round",`
 padding: 0 calc(var(--n-height) / 3) 0 calc(var(--n-height) / 2);
 `)]),y("disabled",`
 cursor: not-allowed !important;
 opacity: var(--n-opacity-disabled);
 `),y("checkable",`
 cursor: pointer;
 box-shadow: none;
 color: var(--n-text-color-checkable);
 background-color: var(--n-color-checkable);
 `,[w("disabled",[H("&:hover","background-color: var(--n-color-hover-checkable);",[w("checked","color: var(--n-text-color-hover-checkable);")]),H("&:active","background-color: var(--n-color-pressed-checkable);",[w("checked","color: var(--n-text-color-pressed-checkable);")])]),y("checked",`
 color: var(--n-text-color-checked);
 background-color: var(--n-color-checked);
 `,[w("disabled",[H("&:hover","background-color: var(--n-color-checked-hover);"),H("&:active","background-color: var(--n-color-checked-pressed);")])])])]),Ee=Object.assign(Object.assign(Object.assign({},M.props),He),{bordered:{type:Boolean,default:void 0},checked:Boolean,checkable:Boolean,strong:Boolean,triggerClickOnClose:Boolean,onClose:[Array,Function],onMouseenter:Function,onMouseleave:Function,"onUpdate:checked":Function,onUpdateChecked:Function,internalCloseFocusable:{type:Boolean,default:!0},internalCloseIsButtonTag:{type:Boolean,default:!0},onCheckedChange:Function}),Te=ye("n-tag"),Ne=A({name:"Tag",props:Ee,slots:Object,setup(e){const c=Ce(null),{mergedBorderedRef:a,mergedClsPrefixRef:u,inlineThemeDisabled:l,mergedRtlRef:h,mergedComponentPropsRef:i}=K(e),o=$(()=>{var r,t;return e.size||((t=(r=i==null?void 0:i.value)===null||r===void 0?void 0:r.Tag)===null||t===void 0?void 0:t.size)||"medium"}),s=M("Tag","-tag",Me,we,e,u);ke(Te,{roundRef:xe(e,"round")});function b(){if(!e.disabled&&e.checkable){const{checked:r,onCheckedChange:t,onUpdateChecked:f,"onUpdate:checked":d}=e;f&&f(!r),d&&d(!r),t&&t(!r)}}function C(r){if(e.triggerClickOnClose||r.stopPropagation(),!e.disabled){const{onClose:t}=e;t&&fe(t,r)}}const g={setTextContent(r){const{value:t}=c;t&&(t.textContent=r)}},B=J("Tag",h,u),P=$(()=>{const{type:r,color:{color:t,textColor:f}={}}=e,d=o.value,{common:{cubicBezierEaseInOut:k},self:{padding:x,closeMargin:S,borderRadius:m,opacityDisabled:E,textColorCheckable:T,textColorHoverCheckable:j,textColorPressedCheckable:_,textColorChecked:O,colorCheckable:W,colorHoverCheckable:F,colorPressedCheckable:q,colorChecked:Q,colorCheckedHover:X,colorCheckedPressed:Y,closeBorderRadius:Z,fontWeightStrong:ee,[p("colorBordered",r)]:oe,[p("closeSize",d)]:re,[p("closeIconSize",d)]:le,[p("fontSize",d)]:ae,[p("height",d)]:G,[p("color",r)]:ne,[p("textColor",r)]:te,[p("border",r)]:ce,[p("closeIconColor",r)]:D,[p("closeIconColorHover",r)]:se,[p("closeIconColorPressed",r)]:ie,[p("closeColorHover",r)]:de,[p("closeColorPressed",r)]:he}}=s.value,R=me(S);return{"--n-font-weight-strong":ee,"--n-avatar-size-override":`calc(${G} - 8px)`,"--n-bezier":k,"--n-border-radius":m,"--n-border":ce,"--n-close-icon-size":le,"--n-close-color-pressed":he,"--n-close-color-hover":de,"--n-close-border-radius":Z,"--n-close-icon-color":D,"--n-close-icon-color-hover":se,"--n-close-icon-color-pressed":ie,"--n-close-icon-color-disabled":D,"--n-close-margin-top":R.top,"--n-close-margin-right":R.right,"--n-close-margin-bottom":R.bottom,"--n-close-margin-left":R.left,"--n-close-size":re,"--n-color":t||(a.value?oe:ne),"--n-color-checkable":W,"--n-color-checked":Q,"--n-color-checked-hover":X,"--n-color-checked-pressed":Y,"--n-color-hover-checkable":F,"--n-color-pressed-checkable":q,"--n-font-size":ae,"--n-height":G,"--n-opacity-disabled":E,"--n-padding":x,"--n-text-color":f||te,"--n-text-color-checkable":T,"--n-text-color-checked":O,"--n-text-color-hover-checkable":j,"--n-text-color-pressed-checkable":_}}),v=l?be("tag",$(()=>{let r="";const{type:t,color:{color:f,textColor:d}={}}=e;return r+=t[0],r+=o.value[0],f&&(r+=`a${L(f)}`),d&&(r+=`b${L(d)}`),a.value&&(r+="c"),r}),P,e):void 0;return Object.assign(Object.assign({},g),{rtlEnabled:B,mergedClsPrefix:u,contentRef:c,mergedBordered:a,handleClick:b,handleCloseClick:C,cssVars:l?void 0:P,themeClass:v==null?void 0:v.themeClass,onRender:v==null?void 0:v.onRender})},render(){var e,c;const{mergedClsPrefix:a,rtlEnabled:u,closable:l,color:{borderColor:h}={},round:i,onRender:o,$slots:s}=this;o==null||o();const b=U(s.avatar,g=>g&&z("div",{class:`${a}-tag__avatar`},g)),C=U(s.icon,g=>g&&z("div",{class:`${a}-tag__icon`},g));return z("div",{class:[`${a}-tag`,this.themeClass,{[`${a}-tag--rtl`]:u,[`${a}-tag--strong`]:this.strong,[`${a}-tag--disabled`]:this.disabled,[`${a}-tag--checkable`]:this.checkable,[`${a}-tag--checked`]:this.checkable&&this.checked,[`${a}-tag--round`]:i,[`${a}-tag--avatar`]:b,[`${a}-tag--icon`]:C,[`${a}-tag--closable`]:l}],style:this.cssVars,onClick:this.handleClick,onMouseenter:this.onMouseenter,onMouseleave:this.onMouseleave},C||b,z("span",{class:`${a}-tag__content`,ref:"contentRef"},(c=(e=this.$slots).default)===null||c===void 0?void 0:c.call(e)),!this.checkable&&l?z(ve,{clsPrefix:a,class:`${a}-tag__close`,disabled:this.disabled,onClick:this.handleCloseClick,focusable:this.internalCloseFocusable,round:i,isButtonTag:this.internalCloseIsButtonTag,absolute:!0}):null,!this.checkable&&this.mergedBordered?z("div",{class:`${a}-tag__border`,style:{borderColor:h}}):null)}});function je(){return ze}const _e={self:je};let N;function Oe(){if(!Se)return!0;if(N===void 0){const e=document.createElement("div");e.style.display="flex",e.style.flexDirection="column",e.style.rowGap="1px",e.appendChild(document.createElement("div")),e.appendChild(document.createElement("div")),document.body.appendChild(e);const c=e.scrollHeight===1;return document.body.removeChild(e),N=c}return N}const We=Object.assign(Object.assign({},M.props),{align:String,justify:{type:String,default:"start"},inline:Boolean,vertical:Boolean,reverse:Boolean,size:[String,Number,Array],wrapItem:{type:Boolean,default:!0},itemClass:String,itemStyle:[String,Object],wrap:{type:Boolean,default:!0},internalUseGap:{type:Boolean,default:void 0}}),Ge=A({name:"Space",props:We,setup(e){const{mergedClsPrefixRef:c,mergedRtlRef:a,mergedComponentPropsRef:u}=K(e),l=$(()=>{var o,s;return e.size||((s=(o=u==null?void 0:u.value)===null||o===void 0?void 0:o.Space)===null||s===void 0?void 0:s.size)||"medium"}),h=M("Space","-space",void 0,_e,e,c),i=J("Space",a,c);return{useGap:Oe(),rtlEnabled:i,mergedClsPrefix:c,margin:$(()=>{const o=l.value;if(Array.isArray(o))return{horizontal:o[0],vertical:o[1]};if(typeof o=="number")return{horizontal:o,vertical:o};const{self:{[p("gap",o)]:s}}=h.value,{row:b,col:C}=Be(s);return{horizontal:V(C),vertical:V(b)}})}},render(){const{vertical:e,reverse:c,align:a,inline:u,justify:l,itemClass:h,itemStyle:i,margin:o,wrap:s,mergedClsPrefix:b,rtlEnabled:C,useGap:g,wrapItem:B,internalUseGap:P}=this,v=Pe($e(this),!1);if(!v.length)return null;const r=`${o.horizontal}px`,t=`${o.horizontal/2}px`,f=`${o.vertical}px`,d=`${o.vertical/2}px`,k=v.length-1,x=l.startsWith("space-");return z("div",{role:"none",class:[`${b}-space`,C&&`${b}-space--rtl`],style:{display:u?"inline-flex":"flex",flexDirection:e&&!c?"column":e&&c?"column-reverse":!e&&c?"row-reverse":"row",justifyContent:["start","end"].includes(l)?`flex-${l}`:l,flexWrap:!s||e?"nowrap":"wrap",marginTop:g||e?"":`-${d}`,marginBottom:g||e?"":`-${d}`,alignItems:a,gap:g?`${o.vertical}px ${o.horizontal}px`:""}},!B&&(g||P)?v:v.map((S,m)=>S.type===Ie?S:z("div",{role:"none",class:h,style:[i,{maxWidth:"100%"},g?"":e?{marginBottom:m!==k?f:""}:C?{marginLeft:x?l==="space-between"&&m===k?"":t:m!==k?r:"",marginRight:x?l==="space-between"&&m===0?"":t:"",paddingTop:d,paddingBottom:d}:{marginRight:x?l==="space-between"&&m===k?"":t:m!==k?r:"",marginLeft:x?l==="space-between"&&m===0?"":t:"",paddingTop:d,paddingBottom:d}]},S)))}});export{Ge as N,Ne as a,$e as g};
