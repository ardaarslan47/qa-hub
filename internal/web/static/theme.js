(() => {
  const storageKey = 'qa-hub-theme';
  const root = document.documentElement;
  const systemDark = window.matchMedia('(prefers-color-scheme: dark)');

  const readChoice = () => {
    try {
      const choice = localStorage.getItem(storageKey);
      return choice === 'light' || choice === 'dark' ? choice : null;
    } catch {
      return null;
    }
  };

  const saveChoice = (theme) => {
    try {
      localStorage.setItem(storageKey, theme);
    } catch {
      return;
    }
  };

  const applySavedChoice = () => {
    const choice = readChoice();
    if (choice) root.dataset.theme = choice;
  };

  const currentTheme = () => root.dataset.theme ?? (systemDark.matches ? 'dark' : 'light');

  applySavedChoice();

  const button = document.createElement('button');
  button.type = 'button';
  button.id = 'theme-toggle';
  button.setAttribute('aria-label', 'Dark theme');

  const showPressedState = () => button.setAttribute('aria-pressed', String(currentTheme() === 'dark'));

  button.addEventListener('click', () => {
    const theme = currentTheme() === 'dark' ? 'light' : 'dark';
    root.dataset.theme = theme;
    saveChoice(theme);
    showPressedState();
  });
  systemDark.addEventListener('change', showPressedState);
  window.addEventListener('pageshow', (event) => {
    if (!event.persisted) return;
    applySavedChoice();
    showPressedState();
  });

  const attachButton = () => {
    showPressedState();
    document.body.prepend(button);
  };
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', attachButton);
  else attachButton();
})();
