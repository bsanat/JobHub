# Accessibility checklist

Use this checklist for UI changes before handing off a JobHub feature.

## Interaction

- Interactive elements are reachable by keyboard.
- Focus order follows the visual flow.
- Buttons use button elements and links use links.
- Disabled states are visually and semantically clear.

## Content

- Form fields have labels or accessible names.
- Error messages identify the field or action that failed.
- Loading and empty states are announced through visible text.
- Icon-only controls have accessible labels.

## Layout

- Text remains readable at narrow widths.
- Long company names, job titles and skill labels wrap without overlapping.
- Kazakh and Russian strings are checked in the same component state.
