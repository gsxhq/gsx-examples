function o(e){let t=0;const n=c=>{t=c,e.innerHTML=`count is ${t}`};e.addEventListener("click",()=>n(t+1)),n(0)}o(document.querySelector("#counter"));
