#!/usr/bin/env python3
"""Builds data/tajweed.json: the tajweed rules of each verse, as ranges of
the Tanzil text in data/quran-uthmani.txt.

    python3 scripts/build_tajweed.py

The rules are Collin Fair's quran-tajweed annotations
(https://github.com/cpfair/quran-tajweed, CC BY 4.0). They are given as
offsets into a 2017 copy of the Tanzil text; Tanzil's current text adds the
pause marks (waqf), the rub' el hizb and sajdah signs, and spells a few
hamzas differently. This script moves each range onto the current text by
aligning the two, and names the rules as Quranku does. That is the only
change made to the annotations.

The output is {"verses": [[[start, end, rule], ...], ...]}, one list a
verse in the order of the text, start and end counting code points of the
verse's text as Tanzil gives it (with the basmalah that opens verse 1 of
most surahs).
"""
import difflib
import json
import os
import urllib.request

HERE = os.path.dirname(os.path.abspath(__file__))
DATA = os.path.join(HERE, "..", "data")
UPSTREAM = "https://raw.githubusercontent.com/cpfair/quran-tajweed/master/output/tajweed.hafs.uthmani-pause-sajdah.json"
# The copy of the Tanzil text the annotations were made on.
UPSTREAM_TEXT = "https://github.com/cpfair/quran-tajweed/files/7281388/quran-uthmani.txt"

# The annotations' rule names, as tajweed.go names them.
RULES = {
    "hamzat_wasl": "ham_wasl",
    "lam_shamsiyyah": "laam_shamsiyah",
    "silent": "slnt",
    "madd_2": "madda_normal",
    "madd_246": "madda_permissible",
    "madd_munfasil": "madda_permissible",
    "madd_muttasil": "madda_obligatory",
    "madd_6": "madda_necessary",
    "ghunnah": "ghunnah",
    "ikhfa": "ikhafa",
    "ikhfa_shafawi": "ikhafa_shafawi",
    "idghaam_ghunnah": "idgham_ghunnah",
    "idghaam_no_ghunnah": "idgham_wo_ghunnah",
    "idghaam_shafawi": "idgham_shafawi",
    "idghaam_mutajanisayn": "idgham_mutajanisayn",
    "idghaam_mutaqaribayn": "idgham_mutaqaribayn",
    "iqlab": "iqlab",
    "qalqalah": "qalaqah",
}


# Signs and spaces never at the edge of a rule: the pause marks (waqf),
# rub' el hizb and sajdah.
EDGE = set(" \u06d6\u06d7\u06d8\u06d9\u06da\u06db\u06de\u06e9")


def get(url):
    req = urllib.request.Request(url, headers={"User-Agent": "quranku-data-script"})
    with urllib.request.urlopen(req, timeout=60) as r:
        return r.read().decode("utf-8")


def verses(text):
    return [line.split("|", 2)[2] for line in text.splitlines() if line[:1].isdigit()]


def mapping(old, new):
    """Maps each offset of old (0..len) to an offset of new."""
    m = [0] * (len(old) + 1)
    sm = difflib.SequenceMatcher(None, old, new, autojunk=False)
    for op, i1, i2, j1, j2 in sm.get_opcodes():
        for k in range(i1, i2 + 1):
            if op == "equal":
                m[k] = j1 + (k - i1)
            elif k == i1:
                m[k] = j1
            else:  # replaced or deleted: the end of the new block
                m[k] = j2
    return m


def main():
    annotations = json.loads(get(UPSTREAM))
    old = verses(get(UPSTREAM_TEXT))
    with open(os.path.join(DATA, "quran-uthmani.txt"), encoding="utf-8") as f:
        new = verses(f.read())
    assert len(annotations) == len(old) == len(new) == 6236
    out, moved = [], 0
    for a, o, n in zip(annotations, old, new):
        m = mapping(o, n) if o != n else None
        ranges = []
        for r in sorted(a["annotations"], key=lambda r: (r["start"], r["end"])):
            s, e = r["start"], r["end"]
            if m:
                s, e = m[s], m[e]
                # A sign the new text adds next to a range is not part of
                # it: trim the pause marks and spaces at its ends.
                while s < e and n[s] in EDGE:
                    s += 1
                while e > s and n[e - 1] in EDGE:
                    e -= 1
                if o[r["start"]:r["end"]] != n[s:e]:
                    moved += 1
            if e > s:
                ranges.append([s, e, RULES[r["rule"]]])
        out.append(ranges)
    with open(os.path.join(DATA, "tajweed.json"), "w", encoding="utf-8") as f:
        json.dump({"source": "cpfair/quran-tajweed (CC BY 4.0), moved onto Tanzil's current Uthmani text",
                   "verses": out}, f, ensure_ascii=False, separators=(",", ":"))
    print(f"{sum(map(len, out))} ranges, {moved} whose text changed with the text")


if __name__ == "__main__":
    main()
