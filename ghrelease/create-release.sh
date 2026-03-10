#!/usr/bin/env sh
set -e

cd /src

# --- Determine the latest semver tag ---
LAST_TAG=$(git tag --list 'v*' --sort=-version:refname | head -n1)

if [ -z "$LAST_TAG" ]; then
  LAST_TAG="v0.0.0"
  RANGE="HEAD"
  echo "No previous tag found, starting from v0.0.0"
else
  RANGE="${LAST_TAG}..HEAD"
  echo "Last tag: ${LAST_TAG}"
fi

# --- Collect commits since the last tag ---
COMMITS=$(git log --pretty=format:"%s" ${RANGE})

if [ -z "$COMMITS" ]; then
  echo "No new commits since ${LAST_TAG}. Nothing to release."
  exit 0
fi

# --- Classify commits and determine bump ---
BUMP="patch"
BREAKING=""
FEATURES=""
FIXES=""
CHORES=""
OTHER=""

IFS='
'
for LINE in $COMMITS; do
  case "$LINE" in
    *"⚠️ breaking:"*|*":warning: breaking:"*|*"breaking:"*)
      BUMP="major"
      BREAKING="${BREAKING}
- ${LINE}"
      ;;
    *"✨ feat:"*|*":sparkles: feat:"*|*"feat:"*)
      if [ "$BUMP" != "major" ]; then
        BUMP="minor"
      fi
      FEATURES="${FEATURES}
- ${LINE}"
      ;;
    *"🔧 fix:"*|*":wrench: fix:"*|*"fix:"*)
      FIXES="${FIXES}
- ${LINE}"
      ;;
    *"🧹 chore:"*|*":broom: chore:"*|*"chore:"*)
      CHORES="${CHORES}
- ${LINE}"
      ;;
    *)
      OTHER="${OTHER}
- ${LINE}"
      ;;
  esac
done

echo "Bump type: ${BUMP}"

# --- Parse the old version and compute the new one ---
VERSION="${LAST_TAG#v}"
MAJOR=$(echo "$VERSION" | cut -d. -f1)
MINOR=$(echo "$VERSION" | cut -d. -f2)
PATCH=$(echo "$VERSION" | cut -d. -f3)

case "$BUMP" in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
esac

NEW_TAG="v${MAJOR}.${MINOR}.${PATCH}"
echo "New tag: ${NEW_TAG}"

# --- Build release notes ---
NOTES=""

if [ -n "$BREAKING" ]; then
  NOTES="${NOTES}## ⚠️ Breaking Changes
${BREAKING}

"
fi

if [ -n "$FEATURES" ]; then
  NOTES="${NOTES}## ✨ Features
${FEATURES}

"
fi

if [ -n "$FIXES" ]; then
  NOTES="${NOTES}## 🔧 Fixes
${FIXES}

"
fi

if [ -n "$CHORES" ]; then
  NOTES="${NOTES}## 🧹 Chores
${CHORES}

"
fi

if [ -n "$OTHER" ]; then
  NOTES="${NOTES}## 📦 Other Changes
${OTHER}

"
fi

echo "---"
echo "$NOTES"
echo "---"

# --- Create the release ---
if [ "$DRY_RUN" = "true" ]; then
  echo "[dry-run] Would create release ${NEW_TAG} in ${GH_REPO}"
else
  echo "$NOTES" | gh release create "$NEW_TAG" \
    --repo "$GH_REPO" \
    --title "$NEW_TAG" \
    --notes-file -
fi

echo "$NEW_TAG"
