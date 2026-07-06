# Lychee Assistant

English | [简体中文](./README.md)

> An AI programming assistant web app. Project name: `lychee`.

## Introduction

I am building an AI programming assistant web app named **Lychee Assistant** (Chinese name **荔枝小秘书**), designed to assist with AI-assisted programming.

## Tech Stack

- **Backend framework**: Gin
- **ORM**: GORM
- **Database**: SQLite
- **Template engine**: Go standard `html/template` and a lightweight Jet template (for rendering frontend pages)

## Features

- **Multilingual support**: Currently supports English and Chinese, switchable via a button on the page (defaults to the system language).
- **Registration**: Sign up with an email address; a verification code is sent to the email for verification during registration.
- **Login**: Log in with email and password using JWT authentication. Each user's data is isolated, with silent token refresh via a dual-token mechanism.
- **Forgot password**: Reset your password via an email verification code.
- **Account settings**: Change your password via an email verification code.
- **Phrase management**: Each user has their own set of common phrases that can be created, viewed, updated, and deleted on the phrase management list page.
- **Task list**: Manage task prompts. The page is split into two panels:
  - **Left — Group management**: Supports grouping, and groups can be nested within other groups.
  - **Right — Task management**: Click any group on the left to display its task list on the right.
    - Task fields: content, created time, updated time, due date, priority (P0–P5), and status (Editing, Pending, Completed).
    - Sortable by created time, updated time, due date, and priority.
    - Action buttons:
      - **Edit**: Open a dialog to modify the task content, priority, and status. While editing, you can pick a common phrase to insert at the cursor position in the content input box.
      - **Copy**: Copy the task content to the clipboard.
- **Check-in (habit tracking)**: Log activities such as drinking water (amount), exercise (duration in minutes), and napping (duration in minutes). View daily check-in reports and customize daily health goals.
  - An **AI health analysis** feature will be added in the future to automatically analyze your health status from check-in data and provide recommendations.
- **Activity logs**: Records activity logs (including login logs). View your own operation history and reports (operations per day and per hour, displayed as bar charts).

## Notes

- **API rate limiting**: Each endpoint is limited to one request per second per IP.
