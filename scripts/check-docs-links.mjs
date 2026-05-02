#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const root = process.cwd();
const files = [path.join(root, "README.md"), ...walk(path.join(root, "docs"))];
const failures = [];

for (const file of files) {
  const markdown = fs.readFileSync(file, "utf8");
  const headings = headingAnchors(markdown);
  const links = markdown.matchAll(/\[[^\]]+\]\(([^)]+)\)/g);
  for (const match of links) {
    const href = match[1].trim();
    if (!href || href.startsWith("http://") || href.startsWith("https://") || href.startsWith("mailto:")) {
      continue;
    }
    const [rawPath, rawAnchor] = href.split("#", 2);
    const linkPath = stripAngleBrackets(rawPath);
    const target = linkPath ? path.resolve(path.dirname(file), linkPath) : file;
    if (!fs.existsSync(target)) {
      failures.push(`${rel(file)} links to missing ${href}`);
      continue;
    }
    if (rawAnchor && target.endsWith(".md")) {
      const targetHeadings = target === file ? headings : headingAnchors(fs.readFileSync(target, "utf8"));
      if (!targetHeadings.has(rawAnchor)) {
        failures.push(`${rel(file)} links to missing heading ${href}`);
      }
    }
  }
}

if (failures.length) {
  console.error(failures.join("\n"));
  process.exit(1);
}

console.log(`checked ${files.length} markdown files: internal links ok`);

function walk(dir) {
  return fs
    .readdirSync(dir, { withFileTypes: true })
    .flatMap((entry) => {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) return walk(full);
      return entry.name.endsWith(".md") ? [full] : [];
    })
    .sort();
}

function headingAnchors(markdown) {
  const anchors = new Set();
  for (const match of markdown.matchAll(/^#{1,6}\s+(.+)$/gm)) {
    anchors.add(slugify(match[1]));
  }
  return anchors;
}

function slugify(text) {
  return text
    .toLowerCase()
    .replace(/`([^`]+)`/g, "$1")
    .replace(/<[^>]+>/g, "")
    .replace(/[^\w\s-]/g, "")
    .trim()
    .replace(/\s+/g, "-");
}

function stripAngleBrackets(text) {
  if (text.startsWith("<") && text.endsWith(">")) return text.slice(1, -1);
  return text;
}

function rel(file) {
  return path.relative(root, file).replaceAll(path.sep, "/");
}
