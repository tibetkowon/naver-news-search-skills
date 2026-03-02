---
name: write-code-tutor
description: Generates an educational review document in docs/reviews/ to explain Go/React code logic and concepts in Korean.
---
# Skill: Write Code Review and Tutor Document

## Description
Acts as a personal coding tutor for gowon, explaining the logic and language-specific concepts of the newly written code in Korean.

## Trigger
- Immediately after completing a significant feature implementation.
- When the user explicitly requests a code review or explanation.

## Instructions
1. **Analyze the Code:** Review the Go backend or React frontend code just modified.
2. **Create Tutor Document:** Create a new markdown file in `docs/reviews/`.
3. **Structure (Korean):**
   - **개요:** 코드의 목적 설명.
   - **상세 해설:** Go의 구조체, 에러 핸들링, 혹은 API 연동 로직 설명.
   - **핵심 요약:** 기억해야 할 디자인 패턴이나 문법 정리.
4. **Report:** Quietly create the file and notify the user: "코드 리뷰 문서가 `docs/reviews/`에 생성되었습니다."
