(this["webpackJsonpfirstcoin-ui"]=this["webpackJsonpfirstcoin-ui"]||[]).push([[0],[,,,,function(e,t,n){"use strict";var c;n.d(t,"a",(function(){return c})),n.d(t,"c",(function(){return r})),n.d(t,"b",(function(){return s})),n.d(t,"d",(function(){return a})),n.d(t,"f",(function(){return i})),n.d(t,"e",(function(){return o})),function(e){e.BLOCKCHAIN="BLOCKCHAIN",e.BLOCK="BLOCK",e.HOST_DETAILS="HOST_DETAILS",e.UNCONFIRMED_TX_POOL="UNCONFIRMED_TX_POOL",e.CLEAR_HOSTS="CLEAR_HOSTS",e.STATUS_MESSAGE="STATUS_MESSAGE"}(c||(c={}));var r=function(e){return{type:c.BLOCKCHAIN,payload:e}},s=function(e){return{type:c.BLOCK,payload:e}},a=function(e){return{type:c.HOST_DETAILS,payload:e}},i=function(e){return{type:c.UNCONFIRMED_TX_POOL,payload:e}},o=function(e){return{type:c.STATUS_MESSAGE,payload:e}}},,,function(e,t,n){"use strict";n.d(t,"a",(function(){return l})),n.d(t,"c",(function(){return u})),n.d(t,"d",(function(){return j})),n.d(t,"b",(function(){return f}));var c=n(2),r=n(3),s=n.n(r),a=n(5),i=n(4),o=Object({NODE_ENV:"production",PUBLIC_URL:"",WDS_SOCKET_HOST:void 0,WDS_SOCKET_PATH:void 0,WDS_SOCKET_PORT:void 0,FAST_REFRESH:!0}).REACT_APP_HOST_NAME?Object({NODE_ENV:"production",PUBLIC_URL:"",WDS_SOCKET_HOST:void 0,WDS_SOCKET_PATH:void 0,WDS_SOCKET_PORT:void 0,FAST_REFRESH:!0}).REACT_APP_HOST_NAME:"",l=function(){var e=Object(a.a)(s.a.mark((function e(){var t;return s.a.wrap((function(e){for(;;)switch(e.prev=e.next){case 0:return e.next=2,fetch("".concat(o,"/block-chain"));case 2:if(!(t=e.sent).ok){e.next=5;break}return e.abrupt("return",t.json());case 5:throw new Error("fetch blockchain returns "+t.status);case 6:case"end":return e.stop()}}),e)})));return function(){return e.apply(this,arguments)}}(),u=function(){var e=Object(a.a)(s.a.mark((function e(){var t,n;return s.a.wrap((function(e){for(;;)switch(e.prev=e.next){case 0:return e.next=2,fetch("".concat(o,"/create-block"),{method:"POST",body:JSON.stringify({a:1,b:"Textual content"})});case 2:if(!(t=e.sent).ok){e.next=5;break}return e.abrupt("return",t.json());case 5:return e.next=7,t.json();case 7:throw n=e.sent,console.log(n),new Error("create blockchain returns "+t.status);case 10:case"end":return e.stop()}}),e)})));return function(){return e.apply(this,arguments)}}(),d=function(){var e=Object(a.a)(s.a.mark((function e(){var t;return s.a.wrap((function(e){for(;;)switch(e.prev=e.next){case 0:return e.next=2,fetch("".concat(o,"/hosts"),{method:"POST",body:JSON.stringify({})});case 2:if(!(t=e.sent).ok){e.next=5;break}return e.abrupt("return",t.json());case 5:throw new Error("fetch hosts returns "+t.status);case 6:case"end":return e.stop()}}),e)})));return function(){return e.apply(this,arguments)}}(),b=function(){var e=Object(a.a)(s.a.mark((function e(){var t;return s.a.wrap((function(e){for(;;)switch(e.prev=e.next){case 0:return e.next=2,fetch("".concat(o,"/txpool"),{method:"GET"});case 2:if(!(t=e.sent).ok){e.next=5;break}return e.abrupt("return",t.json());case 5:throw new Error("fetch host details returns "+t.status);case 6:case"end":return e.stop()}}),e)})));return function(){return e.apply(this,arguments)}}(),j=function(){var e=Object(a.a)(s.a.mark((function e(t,n,c){var r,a;return s.a.wrap((function(e){for(;;)switch(e.prev=e.next){case 0:return e.next=2,fetch("".concat(o,"/spend-coin-relay"),{method:"POST",body:JSON.stringify({host:t,address:n,amount:c})});case 2:if(!(r=e.sent).ok){e.next=5;break}return e.abrupt("return",r.json());case 5:return e.next=7,r.json();case 7:throw a=e.sent,new Error("pay from host ".concat(t," returns ").concat(r.status," ").concat(a.message));case 9:case"end":return e.stop()}}),e)})));return function(t,n,c){return e.apply(this,arguments)}}(),f=function(){var e=Object(a.a)(s.a.mark((function e(t){var n,r;return s.a.wrap((function(e){for(;;)switch(e.prev=e.next){case 0:return e.prev=0,e.next=3,d();case 3:n=e.sent;try{t(Object(i.d)(Object(c.a)({},n)))}catch(s){console.log(s)}return e.next=7,b();case 7:r=e.sent,t(Object(i.f)({pool:Object.values(r)})),e.next=14;break;case 11:e.prev=11,e.t0=e.catch(0),console.log(e.t0);case 14:case"end":return e.stop()}}),e,null,[[0,11]])})));return function(t){return e.apply(this,arguments)}}()},function(e,t,n){"use strict";n.d(t,"a",(function(){return c})),n.d(t,"c",(function(){return a}));var c,r=n(2),s=n(4);!function(e){e.ERROR="ERROR",e.SUCCESS="SUCCESS",e.NOTICE="NOTICE",e.BLANK="BLANK"}(c||(c={}));var a={blockchain:{blocks:[]},hostDetails:[],unconfirmedTxPool:[],statusMessage:{message:""}};t.b=function(e,t){switch(console.log("-------------state before-----------------"),console.log(e),console.log("----------------action--------------------"),console.log(t),t.type){case s.a.BLOCKCHAIN:return Object(r.a)(Object(r.a)({},e),{},{blockchain:Object(r.a)(Object(r.a)({},e.blockchain),t.payload)});case s.a.BLOCK:return Object(r.a)(Object(r.a)({},e),{},{blockchain:Object(r.a)(Object(r.a)({},e.blockchain),t.payload)});case s.a.HOST_DETAILS:var n=JSON.parse(JSON.stringify(t.payload)),c=[];Object.values(n).forEach((function(e){return c.push(e)}));var a=function(e){return parseInt(e.split(":")[1])},i=c.sort((function(e,t){return a(e.hostname)<a(t.hostname)?-1:1}));return Object(r.a)(Object(r.a)({},e),{},{hostDetails:i});case s.a.CLEAR_HOSTS:return Object(r.a)(Object(r.a)({},e),{},{hostDetails:[]});case s.a.UNCONFIRMED_TX_POOL:return Object(r.a)(Object(r.a)({},e),{},{unconfirmedTxPool:t.payload.pool});case s.a.STATUS_MESSAGE:return t.payload.message!==e.statusMessage.message&&t.payload.level!==e.statusMessage.level?Object(r.a)(Object(r.a)({},e),{},{statusMessage:Object(r.a)({},t.payload)}):e;default:return e}}},function(e,t,n){"use strict";n.d(t,"b",(function(){return a}));n(11),n(12);var c=n(0),r=function(e){var t=e.txOData;return Object(c.jsxs)("div",{children:[Object(c.jsxs)("div",{children:["scriptPubKey: ",a(t.scriptPubKey)]}),Object(c.jsxs)("div",{children:["value: ",t.value]})]})},s=function(e){var t=e.txInData;return Object(c.jsxs)("div",{children:[Object(c.jsxs)("div",{children:["txid: ",t.txid]}),Object(c.jsxs)("div",{children:["vout: ",t.vout]}),Object(c.jsxs)("div",{children:["scriptSig: ",t.scriptSig.substring(0,5)," ..."]})]})},a=function(e){return"".concat(e.substring(0,5),"...").concat(e.substring(20,25),"...").concat(e.substring(e.length-6,e.length-1))};t.a=function(e){var t=e.txData;return Object(c.jsxs)("div",{className:"text-xs",children:[Object(c.jsxs)("div",{children:["id: ",t.txid.substring(0,5),"..."]}),Object(c.jsxs)("div",{children:["timestamp: ",t.timestamp.toString().substring(0,10),"..."]}),Object(c.jsxs)("div",{children:["txO: [",t.vout.map((function(e,t){return Object(c.jsx)("div",{className:"pl2",children:Object(c.jsx)(r,{txOData:e})},t)})),"]"]}),Object(c.jsxs)("div",{children:["txIn: [",t.vin.map((function(e,t){return Object(c.jsx)("div",{className:"pl2",children:Object(c.jsx)(s,{txInData:e})},t)})),"]"]})]})}},,,function(e,t,n){},function(e,t,n){},,,,,function(e,t,n){"use strict";(function(e){var c=n(9),r=n(0);t.a=function(t){var n,s,a=t.unconfirmedTxData;return Object(r.jsxs)("div",{className:"border border-black w-11/12 mb-10 bg-white",children:[Object(r.jsxs)("div",{className:"p-2",children:[Object(r.jsx)("span",{className:"font-bold",children:"In tx:"}),Object(c.b)(a.vin[0].txid)]}),Object(r.jsxs)("div",{className:"p-2",children:[Object(r.jsx)("span",{className:"font-bold",children:"pay to:"})," ",Object(c.b)(e.from(a.vout[0].scriptPubKey,"base64").toString("ascii")),"  ",Object(r.jsx)("span",{className:"font-bold",children:"amount: "}),a.vout[0].value]}),Object(r.jsx)("div",{className:"p-2",children:(null===a||void 0===a?void 0:a.vout[1])&&Object(r.jsxs)(r.Fragment,{children:[Object(r.jsx)("span",{className:"font-bold",children:"pay from:"})," ",Object(c.b)(e.from(null===a||void 0===a||null===(n=a.vout[1])||void 0===n?void 0:n.scriptPubKey,"base64").toString("ascii")),"  ",Object(r.jsx)("span",{className:"font-bold",children:"tx out: "}),null===a||void 0===a||null===(s=a.vout[1])||void 0===s?void 0:s.value]})})]})}}).call(this,n(14).Buffer)},function(e,t,n){"use strict";(function(e){var c=n(3),r=n.n(c),s=n(5),a=n(6),i=n(1),o=n(4),l=n(7),u=n(9),d=(n(36),n(8)),b=n(0);t.a=function(t){var n=t.host,c=t.address,j=t.totalAmount,f=t.dispatch,h=Object(i.useState)(""),O=Object(a.a)(h,2),x=O[0],v=O[1],p=Object(i.useState)(0),m=Object(a.a)(p,2),w=m[0],y=m[1],g=function(){var e=Object(s.a)(r.a.mark((function e(){return r.a.wrap((function(e){for(;;)switch(e.prev=e.next){case 0:return e.prev=0,e.next=3,Object(l.d)(n,x,w);case 3:e.sent,Object(l.b)(f),v(""),y(0),e.next=13;break;case 9:e.prev=9,e.t0=e.catch(0),f(Object(o.e)({level:d.a.ERROR,message:e.t0})),console.log(e.t0);case 13:case"end":return e.stop()}}),e,null,[[0,9]])})));return function(){return e.apply(this,arguments)}}();return Object(b.jsxs)("div",{className:"rounded-md text-white p-4 min-h-100 w-4/5 bg-trendyBlue mb-10",children:[Object(b.jsxs)("div",{className:"justify-content flex space-x-4 pb-4",children:[Object(b.jsx)("div",{className:"flex items-center",children:Object(b.jsxs)("div",{children:["address: ",Object(u.b)(e.from(c,"base64").toString("ascii"))]})}),Object(b.jsx)("div",{onClick:function(){navigator.clipboard.writeText(c)},className:"h-10 w-16 bg-trendyYellow text-white flex justify-center items-center transform transition duration-500 hover:scale-105 cursor-pointer font-semibold py-2 px-4 rounded",children:"Copy"})]}),Object(b.jsxs)("div",{className:"mb-4 flex",children:[Object(b.jsxs)("div",{className:"flex align-center space-between",children:[Object(b.jsx)("div",{className:"text-white pr-4 whitespace-nowrap flex flex-col justify-center",children:Object(b.jsx)("div",{className:"",children:"Pay:"})}),Object(b.jsx)("textarea",{value:w||"",onChange:function(e){return y(parseInt(e.target.value))},className:"form-textarea mt-1 mr-2 block w-9/12 border-white overflow-hidden",rows:1,placeholder:"Amount"}),Object(b.jsx)("textarea",{value:x||"",onChange:function(e){return v(e.target.value)},className:"form-textarea mt-1 mr-2 block w-9/12 border-white overflow-hidden",rows:1,placeholder:"To"})]}),Object(b.jsx)("div",{onClick:g,className:"h-10 w-16 mt-1 bg-trendyYellow text-white flex justify-center items-center transform transition duration-500 hover:scale-105 cursor-pointer font-semibold py-2 px-4 rounded",children:"Pay"})]}),Object(b.jsxs)("div",{children:["Balance: ",j]})]})}}).call(this,n(14).Buffer)},,,,,function(e,t,n){},,,,,,,function(e,t,n){},,,,,function(e,t,n){},function(e,t,n){},function(e,t,n){"use strict";n.r(t);var c,r=n(1),s=n.n(r),a=n(15),i=n.n(a),o=(n(24),n(6)),l=n(16),u=n(0),d=function(e){Object(l.a)(e);var t=Object(r.useState)(!1),n=Object(o.a)(t,2),c=n[0],s=n[1];return Object(u.jsx)("div",{children:Object(u.jsxs)("div",{children:[Object(u.jsx)("div",{className:"w-3/6 underline cursor-help whitespace-nowrap",onMouseOver:function(){s(!0)},onMouseOut:function(){s(!1)},children:"How this works?"}),c&&Object(u.jsx)("div",{style:{position:"fixed",zIndex:1e3,height:c?"auto":0},children:Object(u.jsxs)("div",{className:"bg-oliveGreen text-white border-1 rounded border-white p-2 max-w-lg",children:['This is a demonstration blockchain, and it\'s very simple to use! In the middle you will see the blockchain. Click on "Mine" to manually mine a block.',Object(u.jsxs)("div",{className:"mt-3",children:['On the right you will see the "Control Panel". This is where you can transfer coin between wallets. Click "copy" to copy the address you want to pay ',Object(u.jsx)("b",{children:"to"}),'. Select an amount you wish to pay, and click "pay". Only the first miner (wallet) mines which is rewarded 100 coin + 1 coin per transaction.']}),Object(u.jsx)("div",{className:"mt-3",children:'In the control panel you will also see the unconfired transaction pool. It will indicate the payer, payee, transaction id used to confirm the payment, and the "vout" which is the change awarded to the payer.'}),Object(u.jsx)("div",{className:"mt-3",children:"Happy mining!"})]})})]})})},b=function(e){e.state,e.dispatch;return Object(u.jsxs)("div",{className:"fixed z-50 w-8/12 bg-background2 flex flex-col items-center h-24 border-b border-grey mb-10",children:[Object(u.jsx)("div",{className:"text-trendyBlue font-bold text-xl",children:"Firstcoin"}),Object(u.jsxs)("div",{className:"flex space-between w-full",children:[Object(u.jsx)("div",{className:"w-2/6",children:Object(u.jsx)("div",{className:"ml-10",children:Object(u.jsx)(d,{})})}),Object(u.jsx)("div",{className:"flex text-trendyGrey justify-center font-bold w-2/6 whitespace-nowrap",children:"Freddie O'Donnell"}),Object(u.jsx)("a",{href:"https://github.com/FreddieFruitSticks",target:"_blank",className:"flex justify-end w-2/6 text-trendyBlue underline font-bold whitespace-nowrap",children:Object(u.jsx)("div",{className:"mr-10",children:"My Github Profile"})})]}),Object(u.jsx)("div",{className:"text-trendyGrey",children:"Cape Town"})]})},j=n(2),f=n(8),h=s.a.createContext({state:f.c,dispatch:function(e){return console.warn("WARNING! Dispatch function not set. Attemting to dispatch ".concat(e))}}),O=function(e){var t=e.children,n=Object(r.useReducer)(f.b,Object(j.a)(Object(j.a)({},f.c),{})),c=Object(o.a)(n,2),s=c[0],a=c[1];return console.log("-------------state after-----------------"),console.log(s),Object(u.jsx)(h.Provider,{value:{state:s,dispatch:a},children:t})},x=n(3),v=n.n(x),p=n(5),m=(n(11),n(12),n(9)),w=(n(13),n(17)),y=n.n(w);!function(e){e[e.Backwards=0]="Backwards",e[e.Forwards=1]="Forwards"}(c||(c={}));var g=function(e){var t=e.blockData,n=e.last,s=e.direction,a=Object(r.useState)(!1),i=Object(o.a)(a,2),l=i[0],d=i[1];return Object(u.jsxs)("div",{className:"w-full flex flex-col items-center",children:[Object(u.jsx)("div",{className:"w-5/12 transform transition duration-500 hover:scale-105 flex flex-col items-center",children:Object(u.jsx)(y.a,{duration:500,height:l?"auto":150,className:"w-full border-2 border-trendyGrey rounded-lg animate__animated animate__rubberBand",children:Object(u.jsx)("div",{onClick:function(){return d(!l)},className:" cursor-pointer",children:Object(u.jsxs)("div",{className:"rounded text-trendyGrey bg-trendyGreen tran block p-2 overflow-hidden",children:[Object(u.jsxs)("div",{children:["index: ",t.index]}),Object(u.jsxs)("div",{children:["hash: ",t.hash.substring(0,10),"..."]}),Object(u.jsxs)("div",{children:["prev: ",t.previousHash?t.previousHash.substring(0,10):null]}),Object(u.jsxs)("div",{children:["difficulty: ",t.difficultyLevel]}),Object(u.jsxs)("div",{children:["time: ",t.timestamp]}),Object(u.jsxs)("div",{children:["transactions: ",t.transactions.map((function(e,t){return Object(u.jsx)(m.a,{txData:e},t)}))]})]})})})}),s==c.Forwards&&!n&&Object(u.jsx)("span",{className:"arrow arrow-down"})]})},N=n(7),S=n(4),k=(n(31),function(e){var t=e.dispatch,n=Object(r.useState)(!1),c=Object(o.a)(n,2),s=c[0],a=c[1],i=function(){var e=Object(p.a)(v.a.mark((function e(){var n;return v.a.wrap((function(e){for(;;)switch(e.prev=e.next){case 0:return e.prev=0,a(!0),e.next=4,Object(N.c)();case 4:n=e.sent,t(Object(S.b)(n)),Object(N.b)(t),e.next=12;break;case 9:e.prev=9,e.t0=e.catch(0),console.log(e.t0);case 12:a(!1);case 13:case"end":return e.stop()}}),e,null,[[0,9]])})));return function(){return e.apply(this,arguments)}}();return Object(u.jsx)("div",{children:s?Object(u.jsx)("div",{className:"flex justify-center",children:Object(u.jsx)("img",{className:"w-3/6",src:"gntl-mining.gif"})}):Object(u.jsx)("div",{onClick:i,className:"h-24 w-36 mb-28 bg-trendyBlue text-white flex justify-center items-center transform transition duration-500 hover:scale-105 cursor-pointer font-semibold py-2 px-4 rounded",children:"Mine"})})}),T=function(e){var t=e.state,n=e.dispatch;Object(r.useEffect)((function(){function e(){return(e=Object(p.a)(v.a.mark((function e(){var t;return v.a.wrap((function(e){for(;;)switch(e.prev=e.next){case 0:return e.prev=0,e.next=3,Object(N.a)();case 3:t=e.sent,n(Object(S.c)({blocks:t.blocks})),e.next=10;break;case 7:e.prev=7,e.t0=e.catch(0),console.log(e.t0);case 10:case"end":return e.stop()}}),e,null,[[0,7]])})))).apply(this,arguments)}!function(){e.apply(this,arguments)}()}),[n]);var s=Object(r.useRef)(null);return Object(r.useEffect)((function(){var e;null===(e=s.current)||void 0===e||e.scrollIntoView({behavior:"smooth"})}),[t.blockchain]),Object(u.jsxs)("div",{children:[Object(u.jsxs)("div",{className:"grid grid-cols-1 gap-4",children:[Object(u.jsxs)("div",{className:"w-full font-bold flex flex-col items-center",children:["Firstcoin Blockchain",Object(u.jsx)("span",{className:"arrow arrow-down"})]}),t.blockchain.blocks.map((function(e,n){return Object(u.jsx)(g,{direction:c.Forwards,last:n===t.blockchain.blocks.length-1,blockData:e},n)})),Object(u.jsx)("div",{ref:s})]}),Object(u.jsx)("div",{className:"pt-10 flex justify-center items-center w-full",children:Object(u.jsx)(k,{dispatch:n})})]})},E=n(18),C=n(19),_=function(e){var t,n,c=e.state,s=e.dispatch;Object(r.useEffect)((function(){Object(N.b)(s)}),[s]);var a=Object(r.useRef)(null);return Object(r.useEffect)((function(){var e;(null===c||void 0===c?void 0:c.unconfirmedTxPool.length)>0&&(null===(e=a.current)||void 0===e||e.scrollIntoView({behavior:"smooth"}))}),[null===c||void 0===c?void 0:c.unconfirmedTxPool]),Object(u.jsxs)("div",{className:"flex flex-col items-center border-4 border-trendyTurquoise rounded-lg fixed right-0 h-screen bottom-0 w-4/12 overflow-y-scroll",children:[Object(u.jsx)("div",{className:"flex mb-16 items-center justify-center mt-10 font-trendyGrey text-3xl font-medium",children:"Control Panel"}),Object(u.jsx)("div",{className:"flex mb-10 items-center justify-center font-trendyGrey text-xl font-medium",children:"Global Wallets"}),null===c||void 0===c||null===(t=c.hostDetails)||void 0===t?void 0:t.map((function(e,t){return Object(u.jsx)(C.a,{dispatch:s,host:e.hostname,address:e.address,totalAmount:e.totalAmount},t)})),Object(u.jsx)("div",{className:"flex mb-10 items-center justify-center font-trendyGrey text-xl font-medium",children:"Unconfirmed Transaction pool"}),null===c||void 0===c||null===(n=c.unconfirmedTxPool)||void 0===n?void 0:n.map((function(e,t){return Object(u.jsx)(E.a,{unconfirmedTxData:e},t)})),Object(u.jsx)("div",{ref:a})]})},A=(n(37),function(e){var t,n,c=e.state,s=e.dispatch;return Object(r.useEffect)((function(){setTimeout((function(){return s(Object(S.e)({message:"",level:f.a.BLANK}))}),1e4)}),[c.statusMessage]),Object(u.jsx)("div",{children:Object(u.jsx)("div",{style:{position:"fixed",zIndex:1e3,height:c.statusMessage.level?"auto":0},className:"status-message-".concat(c.statusMessage.level," ").concat(c.statusMessage.level&&c.statusMessage.level!=f.a.BLANK?"visible":"invisible"," w-8/12 p-10"),children:(null===c||void 0===c||null===(t=c.statusMessage)||void 0===t?void 0:t.message)&&"".concat(null===c||void 0===c||null===(n=c.statusMessage)||void 0===n?void 0:n.message)})})});var P,D=(P=function(e){var t=e.state,n=e.dispatch;return Object(u.jsxs)("div",{className:"bg-background",children:[Object(u.jsxs)("div",{className:"w-8/12",children:[Object(u.jsx)(A,{state:t,dispatch:n}),Object(u.jsx)(b,{state:t,dispatch:n}),Object(u.jsx)("div",{className:"pt-28",children:Object(u.jsx)(T,{state:t,dispatch:n})})]}),Object(u.jsx)(_,{state:t,dispatch:n})]})},function(e){return Object(u.jsx)(h.Consumer,{children:function(t){var n=t.state,c=t.dispatch;return Object(u.jsx)(P,Object(j.a)(Object(j.a)({},e),{},{state:n,dispatch:c}))}})}),I=function(e){e&&e instanceof Function&&n.e(3).then(n.bind(null,39)).then((function(t){var n=t.getCLS,c=t.getFID,r=t.getFCP,s=t.getLCP,a=t.getTTFB;n(e),c(e),r(e),s(e),a(e)}))};i.a.render(Object(u.jsx)(O,{children:Object(u.jsx)(s.a.StrictMode,{children:Object(u.jsx)(D,{})})}),document.getElementById("root")),I()}],[[38,1,2]]]);
//# sourceMappingURL=main.5765c58d.chunk.js.map