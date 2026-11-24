(async () => {
  const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
  // ChatGPT virtualizes: step-scroll the main container to load all turns.
  const scroller = document.querySelector('main') || document.scrollingElement;
  let lastHeight = 0;
  for (let i = 0; i < 24; i++) {
    scroller.scrollTop = scroller.scrollHeight;
    window.scrollTo(0, document.body.scrollHeight);
    await sleep(350);
    if (scroller.scrollHeight === lastHeight) break;
    lastHeight = scroller.scrollHeight;
  }
  await sleep(600);
  const turns = [...document.querySelectorAll('[data-testid^="conversation-turn"], article')];
  const texts = turns.map((t) => t.innerText || '');
  const joined = texts.join('\n---TURN---\n');
  const idx = joined.lastIndexOf('ROAST REQUEST');
  return JSON.stringify({
    turns: turns.length,
    totalChars: joined.length,
    afterBrief: idx >= 0 ? joined.slice(idx + 2700, idx + 14000) : joined.slice(-11000),
  });
})()