fetch('/healthz').then(r=>r.json()).then(x=>console.log('DRPE',x));
