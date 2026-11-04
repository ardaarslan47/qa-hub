import { expect, test } from './fixtures';
import { createProject, projectList } from './steps';

test('Project oluşturma: ad girilir, project listede görünür, tıklanınca ana sayfasında adı görünür', async ({ page }) => {
  await createProject(page, 'Mobile');

  await expect(projectList(page)).toHaveText(['Mobile']);
  await projectList(page).getByRole('link', { name: 'Mobile' }).click();
  await expect(page.getByRole('heading', { name: 'Mobile' })).toBeVisible();
});

test('Boş ad: project oluşturulamaz, hata mesajı görünür, listeye bir şey eklenmez', async ({ page }) => {
  await createProject(page, 'Mobile');
  await expect(projectList(page)).toHaveText(['Mobile']);

  await createProject(page, '');

  await expect(page.getByRole('alert')).toHaveText('Name is required.');
  await expect(projectList(page)).toHaveText(['Mobile']);
});

test('Aynı ad: Mobile varken mobile oluşturulamaz, hata mesajı görünür, listeye bir şey eklenmez', async ({ page }) => {
  await createProject(page, 'Mobile');
  await expect(projectList(page)).toHaveText(['Mobile']);

  await createProject(page, 'mobile');

  await expect(page.getByRole('alert')).toHaveText('A project with this name already exists.');
  await expect(projectList(page)).toHaveText(['Mobile']);
});
