import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    redirectTo: 'tray',
    pathMatch: 'full'
  },
  {
    path: 'tray',
    loadComponent: () => import('./tray/tray.component').then(m => m.TrayComponent)
  },
  {
    path: 'main',
    loadComponent: () => import('./main/main.component').then(m => m.MainComponent)
  },
  {
    path: 'settings',
    loadComponent: () => import('./settings/settings.component').then(m => m.SettingsComponent)
  },
  {
    path: 'jellyfin',
    loadComponent: () => import('./jellyfin/jellyfin.component').then(m => m.JellyfinComponent)
  }
];
