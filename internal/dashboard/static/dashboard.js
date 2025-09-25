"use strict";
(() => {
  var to = Object.defineProperty;
  var sl = (e, t, r) =>
    t in e
      ? to(e, t, { enumerable: !0, configurable: !0, writable: !0, value: r })
      : (e[t] = r);
  var il = (e, t) => {
    for (var r in t) to(e, r, { get: t[r], enumerable: !0 });
  };
  var ro = (e, t, r) => (sl(e, typeof t != "symbol" ? t + "" : t, r), r);
  var _t,
    L,
    so,
    ll,
    Ue,
    no,
    io,
    lo,
    co,
    Yr,
    Gr,
    Ur,
    uo,
    yt = {},
    fo = [],
    cl = /acit|ex(?:s|g|n|p|$)|rph|grid|ows|mnc|ntw|ine[ch]|zoo|^ord|itera/i,
    wt = Array.isArray;
  function Ne(e, t) {
    for (var r in t) e[r] = t[r];
    return e;
  }
  function Xr(e) {
    e && e.parentNode && e.parentNode.removeChild(e);
  }
  function g(e, t, r) {
    var n,
      a,
      s,
      i = {};
    for (s in t)
      s == "key" ? (n = t[s]) : s == "ref" ? (a = t[s]) : (i[s] = t[s]);
    if (
      (arguments.length > 2 &&
        (i.children = arguments.length > 3 ? _t.call(arguments, 2) : r),
      typeof e == "function" && e.defaultProps != null)
    )
      for (s in e.defaultProps) i[s] === void 0 && (i[s] = e.defaultProps[s]);
    return bt(e, i, n, a, null);
  }
  function bt(e, t, r, n, a) {
    var s = {
      type: e,
      props: t,
      key: r,
      ref: n,
      __k: null,
      __: null,
      __b: 0,
      __e: null,
      __c: null,
      constructor: void 0,
      __v: a ?? ++so,
      __i: -1,
      __u: 0,
    };
    return a == null && L.vnode != null && L.vnode(s), s;
  }
  function Zt() {
    return { current: null };
  }
  function h(e) {
    return e.children;
  }
  function he(e, t) {
    (this.props = e), (this.context = t);
  }
  function nt(e, t) {
    if (t == null) return e.__ ? nt(e.__, e.__i + 1) : null;
    for (var r; t < e.__k.length; t++)
      if ((r = e.__k[t]) != null && r.__e != null) return r.__e;
    return typeof e.type == "function" ? nt(e) : null;
  }
  function mo(e) {
    var t, r;
    if ((e = e.__) != null && e.__c != null) {
      for (e.__e = e.__c.base = null, t = 0; t < e.__k.length; t++)
        if ((r = e.__k[t]) != null && r.__e != null) {
          e.__e = e.__c.base = r.__e;
          break;
        }
      return mo(e);
    }
  }
  function jr(e) {
    ((!e.__d && (e.__d = !0) && Ue.push(e) && !Jt.__r++) ||
      no != L.debounceRendering) &&
      ((no = L.debounceRendering) || io)(Jt);
  }
  function Jt() {
    for (var e, t, r, n, a, s, i, l = 1; Ue.length; )
      Ue.length > l && Ue.sort(lo),
        (e = Ue.shift()),
        (l = Ue.length),
        e.__d &&
          ((r = void 0),
          (n = void 0),
          (a = (n = (t = e).__v).__e),
          (s = []),
          (i = []),
          t.__P &&
            (((r = Ne({}, n)).__v = n.__v + 1),
            L.vnode && L.vnode(r),
            Qr(
              t.__P,
              r,
              n,
              t.__n,
              t.__P.namespaceURI,
              32 & n.__u ? [a] : null,
              s,
              a ?? nt(n),
              !!(32 & n.__u),
              i
            ),
            (r.__v = n.__v),
            (r.__.__k[r.__i] = r),
            go(s, r, i),
            (n.__e = n.__ = null),
            r.__e != a && mo(r)));
    Jt.__r = 0;
  }
  function po(e, t, r, n, a, s, i, l, u, f, m) {
    var c,
      d,
      p,
      x,
      b,
      v,
      y,
      N = (n && n.__k) || fo,
      R = t.length;
    for (u = ul(r, t, N, u, R), c = 0; c < R; c++)
      (p = r.__k[c]) != null &&
        ((d = p.__i == -1 ? yt : N[p.__i] || yt),
        (p.__i = c),
        (v = Qr(e, p, d, a, s, i, l, u, f, m)),
        (x = p.__e),
        p.ref &&
          d.ref != p.ref &&
          (d.ref && Kr(d.ref, null, p), m.push(p.ref, p.__c || x, p)),
        b == null && x != null && (b = x),
        (y = !!(4 & p.__u)) || d.__k === p.__k
          ? (u = ho(p, u, e, y))
          : typeof p.type == "function" && v !== void 0
          ? (u = v)
          : x && (u = x.nextSibling),
        (p.__u &= -7));
    return (r.__e = b), u;
  }
  function ul(e, t, r, n, a) {
    var s,
      i,
      l,
      u,
      f,
      m = r.length,
      c = m,
      d = 0;
    for (e.__k = new Array(a), s = 0; s < a; s++)
      (i = t[s]) != null && typeof i != "boolean" && typeof i != "function"
        ? ((u = s + d),
          ((i = e.__k[s] =
            typeof i == "string" ||
            typeof i == "number" ||
            typeof i == "bigint" ||
            i.constructor == String
              ? bt(null, i, null, null, null)
              : wt(i)
              ? bt(h, { children: i }, null, null, null)
              : i.constructor == null && i.__b > 0
              ? bt(i.type, i.props, i.key, i.ref ? i.ref : null, i.__v)
              : i).__ = e),
          (i.__b = e.__b + 1),
          (l = null),
          (f = i.__i = fl(i, r, u, c)) != -1 &&
            (c--, (l = r[f]) && (l.__u |= 2)),
          l == null || l.__v == null
            ? (f == -1 && (a > m ? d-- : a < m && d++),
              typeof i.type != "function" && (i.__u |= 4))
            : f != u &&
              (f == u - 1
                ? d--
                : f == u + 1
                ? d++
                : (f > u ? d-- : d++, (i.__u |= 4))))
        : (e.__k[s] = null);
    if (c)
      for (s = 0; s < m; s++)
        (l = r[s]) != null &&
          !(2 & l.__u) &&
          (l.__e == n && (n = nt(l)), xo(l, l));
    return n;
  }
  function ho(e, t, r, n) {
    var a, s;
    if (typeof e.type == "function") {
      for (a = e.__k, s = 0; a && s < a.length; s++)
        a[s] && ((a[s].__ = e), (t = ho(a[s], t, r, n)));
      return t;
    }
    e.__e != t &&
      (n &&
        (t && e.type && !t.parentNode && (t = nt(e)),
        r.insertBefore(e.__e, t || null)),
      (t = e.__e));
    do t = t && t.nextSibling;
    while (t != null && t.nodeType == 8);
    return t;
  }
  function Ce(e, t) {
    return (
      (t = t || []),
      e == null ||
        typeof e == "boolean" ||
        (wt(e)
          ? e.some(function (r) {
              Ce(r, t);
            })
          : t.push(e)),
      t
    );
  }
  function fl(e, t, r, n) {
    var a,
      s,
      i,
      l = e.key,
      u = e.type,
      f = t[r],
      m = f != null && (2 & f.__u) == 0;
    if ((f === null && e.key == null) || (m && l == f.key && u == f.type))
      return r;
    if (n > (m ? 1 : 0)) {
      for (a = r - 1, s = r + 1; a >= 0 || s < t.length; )
        if (
          (f = t[(i = a >= 0 ? a-- : s++)]) != null &&
          !(2 & f.__u) &&
          l == f.key &&
          u == f.type
        )
          return i;
    }
    return -1;
  }
  function oo(e, t, r) {
    t[0] == "-"
      ? e.setProperty(t, r ?? "")
      : (e[t] =
          r == null ? "" : typeof r != "number" || cl.test(t) ? r : r + "px");
  }
  function Kt(e, t, r, n, a) {
    var s, i;
    e: if (t == "style")
      if (typeof r == "string") e.style.cssText = r;
      else {
        if ((typeof n == "string" && (e.style.cssText = n = ""), n))
          for (t in n) (r && t in r) || oo(e.style, t, "");
        if (r) for (t in r) (n && r[t] == n[t]) || oo(e.style, t, r[t]);
      }
    else if (t[0] == "o" && t[1] == "n")
      (s = t != (t = t.replace(co, "$1"))),
        (i = t.toLowerCase()),
        (t =
          i in e || t == "onFocusOut" || t == "onFocusIn"
            ? i.slice(2)
            : t.slice(2)),
        e.l || (e.l = {}),
        (e.l[t + s] = r),
        r
          ? n
            ? (r.u = n.u)
            : ((r.u = Yr), e.addEventListener(t, s ? Ur : Gr, s))
          : e.removeEventListener(t, s ? Ur : Gr, s);
    else {
      if (a == "http://www.w3.org/2000/svg")
        t = t.replace(/xlink(H|:h)/, "h").replace(/sName$/, "s");
      else if (
        t != "width" &&
        t != "height" &&
        t != "href" &&
        t != "list" &&
        t != "form" &&
        t != "tabIndex" &&
        t != "download" &&
        t != "rowSpan" &&
        t != "colSpan" &&
        t != "role" &&
        t != "popover" &&
        t in e
      )
        try {
          e[t] = r ?? "";
          break e;
        } catch {}
      typeof r == "function" ||
        (r == null || (r === !1 && t[4] != "-")
          ? e.removeAttribute(t)
          : e.setAttribute(t, t == "popover" && r == 1 ? "" : r));
    }
  }
  function ao(e) {
    return function (t) {
      if (this.l) {
        var r = this.l[t.type + e];
        if (t.t == null) t.t = Yr++;
        else if (t.t < r.u) return;
        return r(L.event ? L.event(t) : t);
      }
    };
  }
  function Qr(e, t, r, n, a, s, i, l, u, f) {
    var m,
      c,
      d,
      p,
      x,
      b,
      v,
      y,
      N,
      R,
      M,
      w,
      _,
      T,
      C,
      V,
      U,
      B = t.type;
    if (t.constructor != null) return null;
    128 & r.__u && ((u = !!(32 & r.__u)), (s = [(l = t.__e = r.__e)])),
      (m = L.__b) && m(t);
    e: if (typeof B == "function")
      try {
        if (
          ((y = t.props),
          (N = "prototype" in B && B.prototype.render),
          (R = (m = B.contextType) && n[m.__c]),
          (M = m ? (R ? R.props.value : m.__) : n),
          r.__c
            ? (v = (c = t.__c = r.__c).__ = c.__E)
            : (N
                ? (t.__c = c = new B(y, M))
                : ((t.__c = c = new he(y, M)),
                  (c.constructor = B),
                  (c.render = ml)),
              R && R.sub(c),
              (c.props = y),
              c.state || (c.state = {}),
              (c.context = M),
              (c.__n = n),
              (d = c.__d = !0),
              (c.__h = []),
              (c._sb = [])),
          N && c.__s == null && (c.__s = c.state),
          N &&
            B.getDerivedStateFromProps != null &&
            (c.__s == c.state && (c.__s = Ne({}, c.__s)),
            Ne(c.__s, B.getDerivedStateFromProps(y, c.__s))),
          (p = c.props),
          (x = c.state),
          (c.__v = t),
          d)
        )
          N &&
            B.getDerivedStateFromProps == null &&
            c.componentWillMount != null &&
            c.componentWillMount(),
            N && c.componentDidMount != null && c.__h.push(c.componentDidMount);
        else {
          if (
            (N &&
              B.getDerivedStateFromProps == null &&
              y !== p &&
              c.componentWillReceiveProps != null &&
              c.componentWillReceiveProps(y, M),
            (!c.__e &&
              c.shouldComponentUpdate != null &&
              c.shouldComponentUpdate(y, c.__s, M) === !1) ||
              t.__v == r.__v)
          ) {
            for (
              t.__v != r.__v &&
                ((c.props = y), (c.state = c.__s), (c.__d = !1)),
                t.__e = r.__e,
                t.__k = r.__k,
                t.__k.some(function (oe) {
                  oe && (oe.__ = t);
                }),
                w = 0;
              w < c._sb.length;
              w++
            )
              c.__h.push(c._sb[w]);
            (c._sb = []), c.__h.length && i.push(c);
            break e;
          }
          c.componentWillUpdate != null && c.componentWillUpdate(y, c.__s, M),
            N &&
              c.componentDidUpdate != null &&
              c.__h.push(function () {
                c.componentDidUpdate(p, x, b);
              });
        }
        if (
          ((c.context = M),
          (c.props = y),
          (c.__P = e),
          (c.__e = !1),
          (_ = L.__r),
          (T = 0),
          N)
        ) {
          for (
            c.state = c.__s,
              c.__d = !1,
              _ && _(t),
              m = c.render(c.props, c.state, c.context),
              C = 0;
            C < c._sb.length;
            C++
          )
            c.__h.push(c._sb[C]);
          c._sb = [];
        } else
          do
            (c.__d = !1),
              _ && _(t),
              (m = c.render(c.props, c.state, c.context)),
              (c.state = c.__s);
          while (c.__d && ++T < 25);
        (c.state = c.__s),
          c.getChildContext != null && (n = Ne(Ne({}, n), c.getChildContext())),
          N &&
            !d &&
            c.getSnapshotBeforeUpdate != null &&
            (b = c.getSnapshotBeforeUpdate(p, x)),
          (V = m),
          m != null &&
            m.type === h &&
            m.key == null &&
            (V = vo(m.props.children)),
          (l = po(e, wt(V) ? V : [V], t, r, n, a, s, i, l, u, f)),
          (c.base = t.__e),
          (t.__u &= -161),
          c.__h.length && i.push(c),
          v && (c.__E = c.__ = null);
      } catch (oe) {
        if (((t.__v = null), u || s != null))
          if (oe.then) {
            for (
              t.__u |= u ? 160 : 128;
              l && l.nodeType == 8 && l.nextSibling;

            )
              l = l.nextSibling;
            (s[s.indexOf(l)] = null), (t.__e = l);
          } else {
            for (U = s.length; U--; ) Xr(s[U]);
            Wr(t);
          }
        else (t.__e = r.__e), (t.__k = r.__k), oe.then || Wr(t);
        L.__e(oe, t, r);
      }
    else
      s == null && t.__v == r.__v
        ? ((t.__k = r.__k), (t.__e = r.__e))
        : (l = t.__e = dl(r.__e, t, r, n, a, s, i, u, f));
    return (m = L.diffed) && m(t), 128 & t.__u ? void 0 : l;
  }
  function Wr(e) {
    e && e.__c && (e.__c.__e = !0), e && e.__k && e.__k.forEach(Wr);
  }
  function go(e, t, r) {
    for (var n = 0; n < r.length; n++) Kr(r[n], r[++n], r[++n]);
    L.__c && L.__c(t, e),
      e.some(function (a) {
        try {
          (e = a.__h),
            (a.__h = []),
            e.some(function (s) {
              s.call(a);
            });
        } catch (s) {
          L.__e(s, a.__v);
        }
      });
  }
  function vo(e) {
    return typeof e != "object" || e == null || (e.__b && e.__b > 0)
      ? e
      : wt(e)
      ? e.map(vo)
      : Ne({}, e);
  }
  function dl(e, t, r, n, a, s, i, l, u) {
    var f,
      m,
      c,
      d,
      p,
      x,
      b,
      v = r.props,
      y = t.props,
      N = t.type;
    if (
      (N == "svg"
        ? (a = "http://www.w3.org/2000/svg")
        : N == "math"
        ? (a = "http://www.w3.org/1998/Math/MathML")
        : a || (a = "http://www.w3.org/1999/xhtml"),
      s != null)
    ) {
      for (f = 0; f < s.length; f++)
        if (
          (p = s[f]) &&
          "setAttribute" in p == !!N &&
          (N ? p.localName == N : p.nodeType == 3)
        ) {
          (e = p), (s[f] = null);
          break;
        }
    }
    if (e == null) {
      if (N == null) return document.createTextNode(y);
      (e = document.createElementNS(a, N, y.is && y)),
        l && (L.__m && L.__m(t, s), (l = !1)),
        (s = null);
    }
    if (N == null) v === y || (l && e.data == y) || (e.data = y);
    else {
      if (
        ((s = s && _t.call(e.childNodes)), (v = r.props || yt), !l && s != null)
      )
        for (v = {}, f = 0; f < e.attributes.length; f++)
          v[(p = e.attributes[f]).name] = p.value;
      for (f in v)
        if (((p = v[f]), f != "children")) {
          if (f == "dangerouslySetInnerHTML") c = p;
          else if (!(f in y)) {
            if (
              (f == "value" && "defaultValue" in y) ||
              (f == "checked" && "defaultChecked" in y)
            )
              continue;
            Kt(e, f, null, p, a);
          }
        }
      for (f in y)
        (p = y[f]),
          f == "children"
            ? (d = p)
            : f == "dangerouslySetInnerHTML"
            ? (m = p)
            : f == "value"
            ? (x = p)
            : f == "checked"
            ? (b = p)
            : (l && typeof p != "function") ||
              v[f] === p ||
              Kt(e, f, p, v[f], a);
      if (m)
        l ||
          (c && (m.__html == c.__html || m.__html == e.innerHTML)) ||
          (e.innerHTML = m.__html),
          (t.__k = []);
      else if (
        (c && (e.innerHTML = ""),
        po(
          t.type == "template" ? e.content : e,
          wt(d) ? d : [d],
          t,
          r,
          n,
          N == "foreignObject" ? "http://www.w3.org/1999/xhtml" : a,
          s,
          i,
          s ? s[0] : r.__k && nt(r, 0),
          l,
          u
        ),
        s != null)
      )
        for (f = s.length; f--; ) Xr(s[f]);
      l ||
        ((f = "value"),
        N == "progress" && x == null
          ? e.removeAttribute("value")
          : x != null &&
            (x !== e[f] ||
              (N == "progress" && !x) ||
              (N == "option" && x != v[f])) &&
            Kt(e, f, x, v[f], a),
        (f = "checked"),
        b != null && b != e[f] && Kt(e, f, b, v[f], a));
    }
    return e;
  }
  function Kr(e, t, r) {
    try {
      if (typeof e == "function") {
        var n = typeof e.__u == "function";
        n && e.__u(), (n && t == null) || (e.__u = e(t));
      } else e.current = t;
    } catch (a) {
      L.__e(a, r);
    }
  }
  function xo(e, t, r) {
    var n, a;
    if (
      (L.unmount && L.unmount(e),
      (n = e.ref) && ((n.current && n.current != e.__e) || Kr(n, null, t)),
      (n = e.__c) != null)
    ) {
      if (n.componentWillUnmount)
        try {
          n.componentWillUnmount();
        } catch (s) {
          L.__e(s, t);
        }
      n.base = n.__P = null;
    }
    if ((n = e.__k))
      for (a = 0; a < n.length; a++)
        n[a] && xo(n[a], t, r || typeof e.type != "function");
    r || Xr(e.__e), (e.__c = e.__ = e.__e = void 0);
  }
  function ml(e, t, r) {
    return this.constructor(e, r);
  }
  function Oe(e, t, r) {
    var n, a, s, i;
    t == document && (t = document.documentElement),
      L.__ && L.__(e, t),
      (a = (n = typeof r == "function") ? null : (r && r.__k) || t.__k),
      (s = []),
      (i = []),
      Qr(
        t,
        (e = ((!n && r) || t).__k = g(h, null, [e])),
        a || yt,
        yt,
        t.namespaceURI,
        !n && r ? [r] : a ? null : t.firstChild ? _t.call(t.childNodes) : null,
        s,
        !n && r ? r : a ? a.__e : t.firstChild,
        n,
        i
      ),
      go(s, e, i);
  }
  function Jr(e, t) {
    Oe(e, t, Jr);
  }
  function bo(e, t, r) {
    var n,
      a,
      s,
      i,
      l = Ne({}, e.props);
    for (s in (e.type && e.type.defaultProps && (i = e.type.defaultProps), t))
      s == "key"
        ? (n = t[s])
        : s == "ref"
        ? (a = t[s])
        : (l[s] = t[s] === void 0 && i != null ? i[s] : t[s]);
    return (
      arguments.length > 2 &&
        (l.children = arguments.length > 3 ? _t.call(arguments, 2) : r),
      bt(e.type, l, n || e.key, a || e.ref, null)
    );
  }
  function Ee(e) {
    function t(r) {
      var n, a;
      return (
        this.getChildContext ||
          ((n = new Set()),
          ((a = {})[t.__c] = this),
          (this.getChildContext = function () {
            return a;
          }),
          (this.componentWillUnmount = function () {
            n = null;
          }),
          (this.shouldComponentUpdate = function (s) {
            this.props.value != s.value &&
              n.forEach(function (i) {
                (i.__e = !0), jr(i);
              });
          }),
          (this.sub = function (s) {
            n.add(s);
            var i = s.componentWillUnmount;
            s.componentWillUnmount = function () {
              n && n.delete(s), i && i.call(s);
            };
          })),
        r.children
      );
    }
    return (
      (t.__c = "__cC" + uo++),
      (t.__ = e),
      (t.Provider =
        t.__l =
        (t.Consumer = function (r, n) {
          return r.children(n);
        }).contextType =
          t),
      t
    );
  }
  (_t = fo.slice),
    (L = {
      __e: function (e, t, r, n) {
        for (var a, s, i; (t = t.__); )
          if ((a = t.__c) && !a.__)
            try {
              if (
                ((s = a.constructor) &&
                  s.getDerivedStateFromError != null &&
                  (a.setState(s.getDerivedStateFromError(e)), (i = a.__d)),
                a.componentDidCatch != null &&
                  (a.componentDidCatch(e, n || {}), (i = a.__d)),
                i)
              )
                return (a.__E = a);
            } catch (l) {
              e = l;
            }
        throw e;
      },
    }),
    (so = 0),
    (ll = function (e) {
      return e != null && e.constructor == null;
    }),
    (he.prototype.setState = function (e, t) {
      var r;
      (r =
        this.__s != null && this.__s != this.state
          ? this.__s
          : (this.__s = Ne({}, this.state))),
        typeof e == "function" && (e = e(Ne({}, r), this.props)),
        e && Ne(r, e),
        e != null && this.__v && (t && this._sb.push(t), jr(this));
    }),
    (he.prototype.forceUpdate = function (e) {
      this.__v && ((this.__e = !0), e && this.__h.push(e), jr(this));
    }),
    (he.prototype.render = h),
    (Ue = []),
    (io =
      typeof Promise == "function"
        ? Promise.prototype.then.bind(Promise.resolve())
        : setTimeout),
    (lo = function (e, t) {
      return e.__v.__b - t.__v.__b;
    }),
    (Jt.__r = 0),
    (co = /(PointerCapture)$|Capture$/i),
    (Yr = 0),
    (Gr = ao(!1)),
    (Ur = ao(!0)),
    (uo = 0);
  var Me,
    G,
    Zr,
    yo,
    ot = 0,
    So = [],
    J = L,
    _o = J.__b,
    wo = J.__r,
    No = J.diffed,
    Co = J.__c,
    To = J.unmount,
    Ro = J.__;
  function je(e, t) {
    J.__h && J.__h(G, e, ot || t), (ot = 0);
    var r = G.__H || (G.__H = { __: [], __h: [] });
    return e >= r.__.length && r.__.push({}), r.__[e];
  }
  function P(e) {
    return (ot = 1), He(Eo, e);
  }
  function He(e, t, r) {
    var n = je(Me++, 2);
    if (
      ((n.t = e),
      !n.__c &&
        ((n.__ = [
          r ? r(t) : Eo(void 0, t),
          function (l) {
            var u = n.__N ? n.__N[0] : n.__[0],
              f = n.t(u, l);
            u !== f && ((n.__N = [f, n.__[1]]), n.__c.setState({}));
          },
        ]),
        (n.__c = G),
        !G.__f))
    ) {
      var a = function (l, u, f) {
        if (!n.__c.__H) return !0;
        var m = n.__c.__H.__.filter(function (d) {
          return !!d.__c;
        });
        if (
          m.every(function (d) {
            return !d.__N;
          })
        )
          return !s || s.call(this, l, u, f);
        var c = n.__c.props !== l;
        return (
          m.forEach(function (d) {
            if (d.__N) {
              var p = d.__[0];
              (d.__ = d.__N), (d.__N = void 0), p !== d.__[0] && (c = !0);
            }
          }),
          (s && s.call(this, l, u, f)) || c
        );
      };
      G.__f = !0;
      var s = G.shouldComponentUpdate,
        i = G.componentWillUpdate;
      (G.componentWillUpdate = function (l, u, f) {
        if (this.__e) {
          var m = s;
          (s = void 0), a(l, u, f), (s = m);
        }
        i && i.call(this, l, u, f);
      }),
        (G.shouldComponentUpdate = a);
    }
    return n.__N || n.__;
  }
  function q(e, t) {
    var r = je(Me++, 3);
    !J.__s && tn(r.__H, t) && ((r.__ = e), (r.u = t), G.__H.__h.push(r));
  }
  function Pe(e, t) {
    var r = je(Me++, 4);
    !J.__s && tn(r.__H, t) && ((r.__ = e), (r.u = t), G.__h.push(r));
  }
  function $(e) {
    return (
      (ot = 5),
      Z(function () {
        return { current: e };
      }, [])
    );
  }
  function tr(e, t, r) {
    (ot = 6),
      Pe(
        function () {
          if (typeof e == "function") {
            var n = e(t());
            return function () {
              e(null), n && typeof n == "function" && n();
            };
          }
          if (e)
            return (
              (e.current = t()),
              function () {
                return (e.current = null);
              }
            );
        },
        r == null ? r : r.concat(e)
      );
  }
  function Z(e, t) {
    var r = je(Me++, 7);
    return tn(r.__H, t) && ((r.__ = e()), (r.__H = t), (r.__h = e)), r.__;
  }
  function ce(e, t) {
    return (
      (ot = 8),
      Z(function () {
        return e;
      }, t)
    );
  }
  function Fe(e) {
    var t = G.context[e.__c],
      r = je(Me++, 9);
    return (
      (r.c = e),
      t ? (r.__ == null && ((r.__ = !0), t.sub(G)), t.props.value) : e.__
    );
  }
  function rr(e, t) {
    J.useDebugValue && J.useDebugValue(t ? t(e) : e);
  }
  function pl(e) {
    var t = je(Me++, 10),
      r = P();
    return (
      (t.__ = e),
      G.componentDidCatch ||
        (G.componentDidCatch = function (n, a) {
          t.__ && t.__(n, a), r[1](n);
        }),
      [
        r[0],
        function () {
          r[1](void 0);
        },
      ]
    );
  }
  function nr() {
    var e = je(Me++, 11);
    if (!e.__) {
      for (var t = G.__v; t !== null && !t.__m && t.__ !== null; ) t = t.__;
      var r = t.__m || (t.__m = [0, 0]);
      e.__ = "P" + r[0] + "-" + r[1]++;
    }
    return e.__;
  }
  function hl() {
    for (var e; (e = So.shift()); )
      if (e.__P && e.__H)
        try {
          e.__H.__h.forEach(er), e.__H.__h.forEach(en), (e.__H.__h = []);
        } catch (t) {
          (e.__H.__h = []), J.__e(t, e.__v);
        }
  }
  (J.__b = function (e) {
    (G = null), _o && _o(e);
  }),
    (J.__ = function (e, t) {
      e && t.__k && t.__k.__m && (e.__m = t.__k.__m), Ro && Ro(e, t);
    }),
    (J.__r = function (e) {
      wo && wo(e), (Me = 0);
      var t = (G = e.__c).__H;
      t &&
        (Zr === G
          ? ((t.__h = []),
            (G.__h = []),
            t.__.forEach(function (r) {
              r.__N && (r.__ = r.__N), (r.u = r.__N = void 0);
            }))
          : (t.__h.forEach(er), t.__h.forEach(en), (t.__h = []), (Me = 0))),
        (Zr = G);
    }),
    (J.diffed = function (e) {
      No && No(e);
      var t = e.__c;
      t &&
        t.__H &&
        (t.__H.__h.length &&
          ((So.push(t) !== 1 && yo === J.requestAnimationFrame) ||
            ((yo = J.requestAnimationFrame) || gl)(hl)),
        t.__H.__.forEach(function (r) {
          r.u && (r.__H = r.u), (r.u = void 0);
        })),
        (Zr = G = null);
    }),
    (J.__c = function (e, t) {
      t.some(function (r) {
        try {
          r.__h.forEach(er),
            (r.__h = r.__h.filter(function (n) {
              return !n.__ || en(n);
            }));
        } catch (n) {
          t.some(function (a) {
            a.__h && (a.__h = []);
          }),
            (t = []),
            J.__e(n, r.__v);
        }
      }),
        Co && Co(e, t);
    }),
    (J.unmount = function (e) {
      To && To(e);
      var t,
        r = e.__c;
      r &&
        r.__H &&
        (r.__H.__.forEach(function (n) {
          try {
            er(n);
          } catch (a) {
            t = a;
          }
        }),
        (r.__H = void 0),
        t && J.__e(t, r.__v));
    });
  var ko = typeof requestAnimationFrame == "function";
  function gl(e) {
    var t,
      r = function () {
        clearTimeout(n), ko && cancelAnimationFrame(t), setTimeout(e);
      },
      n = setTimeout(r, 35);
    ko && (t = requestAnimationFrame(r));
  }
  function er(e) {
    var t = G,
      r = e.__c;
    typeof r == "function" && ((e.__c = void 0), r()), (G = t);
  }
  function en(e) {
    var t = G;
    (e.__c = e.__()), (G = t);
  }
  function tn(e, t) {
    return (
      !e ||
      e.length !== t.length ||
      t.some(function (r, n) {
        return r !== e[n];
      })
    );
  }
  function Eo(e, t) {
    return typeof t == "function" ? t(e) : t;
  }
  var rn = class {
      constructor() {
        ro(this, "baseURL", "/__viz/api");
      }
      async getRequests() {
        let t = await fetch(`${this.baseURL}/requests`);
        if (!t.ok) throw new Error("Failed to fetch requests");
        return t.json();
      }
      async clearRequests() {
        if (!(await fetch(`${this.baseURL}/clear`, { method: "POST" })).ok)
          throw new Error("Failed to clear requests");
      }
      async compareRequests(t) {
        let r = t.map((a) => `id=${a}`).join("&"),
          n = await fetch(`${this.baseURL}/compare?${r}`);
        if (!n.ok) throw new Error("Failed to compare requests");
        return n.json();
      }
      async replayRequest(t) {
        let r = await fetch(`${this.baseURL}/replay`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(t),
        });
        if (!r.ok) throw new Error("Failed to replay request");
        return r.json();
      }
      async getMetrics(t) {
        let r = await fetch(`${this.baseURL}/metrics?id=${t}`);
        if (!r.ok) throw new Error("Failed to fetch metrics");
        return r.json();
      }
      async getFlameGraph(t) {
        let r = await fetch(`${this.baseURL}/flamegraph?id=${t}`);
        if (!r.ok) throw new Error("Failed to fetch flame graph");
        return r.json();
      }
      async getBottlenecks() {
        let t = await fetch(`${this.baseURL}/bottlenecks`);
        if (!t.ok) throw new Error("Failed to fetch bottlenecks");
        return t.json();
      }
      async getSystemInfo() {
        let t = await fetch(`${this.baseURL}/system-info`);
        if (!t.ok) throw new Error("Failed to fetch system info");
        return t.json();
      }
      subscribeToEvents(t) {
        let r = new EventSource(`${this.baseURL}/events`);
        return (
          (r.onmessage = (n) => {
            try {
              let a = JSON.parse(n.data);
              t(a);
            } catch (a) {
              console.error("Failed to parse event data:", a);
            }
          }),
          (r.onerror = (n) => {
            console.error("EventSource error:", n);
          }),
          r
        );
      }
      exportRequests(t) {
        return JSON.stringify(t, null, 2);
      }
      importRequests(t) {
        try {
          let r = JSON.parse(t);
          if (Array.isArray(r)) return r;
          throw new Error("Invalid format: expected an array of requests");
        } catch (r) {
          throw new Error(`Failed to import requests: ${r.message}`);
        }
      }
    },
    fe = new rn();
  function Mo(e) {
    var t,
      r,
      n = "";
    if (typeof e == "string" || typeof e == "number") n += e;
    else if (typeof e == "object")
      if (Array.isArray(e)) {
        var a = e.length;
        for (t = 0; t < a; t++)
          e[t] && (r = Mo(e[t])) && (n && (n += " "), (n += r));
      } else for (r in e) e[r] && (n && (n += " "), (n += r));
    return n;
  }
  function or() {
    for (var e, t, r = 0, n = "", a = arguments.length; r < a; r++)
      (e = arguments[r]) && (t = Mo(e)) && (n && (n += " "), (n += t));
    return n;
  }
  var cn = "-",
    vl = (e) => {
      let t = bl(e),
        { conflictingClassGroups: r, conflictingClassGroupModifiers: n } = e;
      return {
        getClassGroupId: (i) => {
          let l = i.split(cn);
          return l[0] === "" && l.length !== 1 && l.shift(), Lo(l, t) || xl(i);
        },
        getConflictingClassGroupIds: (i, l) => {
          let u = r[i] || [];
          return l && n[i] ? [...u, ...n[i]] : u;
        },
      };
    },
    Lo = (e, t) => {
      if (e.length === 0) return t.classGroupId;
      let r = e[0],
        n = t.nextPart.get(r),
        a = n ? Lo(e.slice(1), n) : void 0;
      if (a) return a;
      if (t.validators.length === 0) return;
      let s = e.join(cn);
      return t.validators.find(({ validator: i }) => i(s))?.classGroupId;
    },
    Po = /^\[(.+)\]$/,
    xl = (e) => {
      if (Po.test(e)) {
        let t = Po.exec(e)[1],
          r = t?.substring(0, t.indexOf(":"));
        if (r) return "arbitrary.." + r;
      }
    },
    bl = (e) => {
      let { theme: t, classGroups: r } = e,
        n = { nextPart: new Map(), validators: [] };
      for (let a in r) an(r[a], n, a, t);
      return n;
    },
    an = (e, t, r, n) => {
      e.forEach((a) => {
        if (typeof a == "string") {
          let s = a === "" ? t : Io(t, a);
          s.classGroupId = r;
          return;
        }
        if (typeof a == "function") {
          if (yl(a)) {
            an(a(n), t, r, n);
            return;
          }
          t.validators.push({ validator: a, classGroupId: r });
          return;
        }
        Object.entries(a).forEach(([s, i]) => {
          an(i, Io(t, s), r, n);
        });
      });
    },
    Io = (e, t) => {
      let r = e;
      return (
        t.split(cn).forEach((n) => {
          r.nextPart.has(n) ||
            r.nextPart.set(n, { nextPart: new Map(), validators: [] }),
            (r = r.nextPart.get(n));
        }),
        r
      );
    },
    yl = (e) => e.isThemeGetter,
    _l = (e) => {
      if (e < 1) return { get: () => {}, set: () => {} };
      let t = 0,
        r = new Map(),
        n = new Map(),
        a = (s, i) => {
          r.set(s, i), t++, t > e && ((t = 0), (n = r), (r = new Map()));
        };
      return {
        get(s) {
          let i = r.get(s);
          if (i !== void 0) return i;
          if ((i = n.get(s)) !== void 0) return a(s, i), i;
        },
        set(s, i) {
          r.has(s) ? r.set(s, i) : a(s, i);
        },
      };
    },
    sn = "!",
    ln = ":",
    wl = ln.length,
    Nl = (e) => {
      let { prefix: t, experimentalParseClassName: r } = e,
        n = (a) => {
          let s = [],
            i = 0,
            l = 0,
            u = 0,
            f;
          for (let x = 0; x < a.length; x++) {
            let b = a[x];
            if (i === 0 && l === 0) {
              if (b === ln) {
                s.push(a.slice(u, x)), (u = x + wl);
                continue;
              }
              if (b === "/") {
                f = x;
                continue;
              }
            }
            b === "["
              ? i++
              : b === "]"
              ? i--
              : b === "("
              ? l++
              : b === ")" && l--;
          }
          let m = s.length === 0 ? a : a.substring(u),
            c = Cl(m),
            d = c !== m,
            p = f && f > u ? f - u : void 0;
          return {
            modifiers: s,
            hasImportantModifier: d,
            baseClassName: c,
            maybePostfixModifierPosition: p,
          };
        };
      if (t) {
        let a = t + ln,
          s = n;
        n = (i) =>
          i.startsWith(a)
            ? s(i.substring(a.length))
            : {
                isExternal: !0,
                modifiers: [],
                hasImportantModifier: !1,
                baseClassName: i,
                maybePostfixModifierPosition: void 0,
              };
      }
      if (r) {
        let a = n;
        n = (s) => r({ className: s, parseClassName: a });
      }
      return n;
    },
    Cl = (e) =>
      e.endsWith(sn)
        ? e.substring(0, e.length - 1)
        : e.startsWith(sn)
        ? e.substring(1)
        : e,
    Tl = (e) => {
      let t = Object.fromEntries(e.orderSensitiveModifiers.map((n) => [n, !0]));
      return (n) => {
        if (n.length <= 1) return n;
        let a = [],
          s = [];
        return (
          n.forEach((i) => {
            i[0] === "[" || t[i]
              ? (a.push(...s.sort(), i), (s = []))
              : s.push(i);
          }),
          a.push(...s.sort()),
          a
        );
      };
    },
    Rl = (e) => ({
      cache: _l(e.cacheSize),
      parseClassName: Nl(e),
      sortModifiers: Tl(e),
      ...vl(e),
    }),
    kl = /\s+/,
    Sl = (e, t) => {
      let {
          parseClassName: r,
          getClassGroupId: n,
          getConflictingClassGroupIds: a,
          sortModifiers: s,
        } = t,
        i = [],
        l = e.trim().split(kl),
        u = "";
      for (let f = l.length - 1; f >= 0; f -= 1) {
        let m = l[f],
          {
            isExternal: c,
            modifiers: d,
            hasImportantModifier: p,
            baseClassName: x,
            maybePostfixModifierPosition: b,
          } = r(m);
        if (c) {
          u = m + (u.length > 0 ? " " + u : u);
          continue;
        }
        let v = !!b,
          y = n(v ? x.substring(0, b) : x);
        if (!y) {
          if (!v) {
            u = m + (u.length > 0 ? " " + u : u);
            continue;
          }
          if (((y = n(x)), !y)) {
            u = m + (u.length > 0 ? " " + u : u);
            continue;
          }
          v = !1;
        }
        let N = s(d).join(":"),
          R = p ? N + sn : N,
          M = R + y;
        if (i.includes(M)) continue;
        i.push(M);
        let w = a(y, v);
        for (let _ = 0; _ < w.length; ++_) {
          let T = w[_];
          i.push(R + T);
        }
        u = m + (u.length > 0 ? " " + u : u);
      }
      return u;
    };
  function El() {
    let e = 0,
      t,
      r,
      n = "";
    for (; e < arguments.length; )
      (t = arguments[e++]) && (r = Oo(t)) && (n && (n += " "), (n += r));
    return n;
  }
  var Oo = (e) => {
    if (typeof e == "string") return e;
    let t,
      r = "";
    for (let n = 0; n < e.length; n++)
      e[n] && (t = Oo(e[n])) && (r && (r += " "), (r += t));
    return r;
  };
  function Ml(e, ...t) {
    let r,
      n,
      a,
      s = i;
    function i(u) {
      let f = t.reduce((m, c) => c(m), e());
      return (r = Rl(f)), (n = r.cache.get), (a = r.cache.set), (s = l), l(u);
    }
    function l(u) {
      let f = n(u);
      if (f) return f;
      let m = Sl(u, r);
      return a(u, m), m;
    }
    return function () {
      return s(El.apply(null, arguments));
    };
  }
  var re = (e) => {
      let t = (r) => r[e] || [];
      return (t.isThemeGetter = !0), t;
    },
    Ho = /^\[(?:(\w[\w-]*):)?(.+)\]$/i,
    Fo = /^\((?:(\w[\w-]*):)?(.+)\)$/i,
    Pl = /^\d+\/\d+$/,
    Il = /^(\d+(\.\d+)?)?(xs|sm|md|lg|xl)$/,
    Al =
      /\d+(%|px|r?em|[sdl]?v([hwib]|min|max)|pt|pc|in|cm|mm|cap|ch|ex|r?lh|cq(w|h|i|b|min|max))|\b(calc|min|max|clamp)\(.+\)|^0$/,
    Dl = /^(rgba?|hsla?|hwb|(ok)?(lab|lch)|color-mix)\(.+\)$/,
    Ll = /^(inset_)?-?((\d+)?\.?(\d+)[a-z]+|0)_-?((\d+)?\.?(\d+)[a-z]+|0)/,
    Ol =
      /^(url|image|image-set|cross-fade|element|(repeating-)?(linear|radial|conic)-gradient)\(.+\)$/,
    at = (e) => Pl.test(e),
    F = (e) => !!e && !Number.isNaN(Number(e)),
    Be = (e) => !!e && Number.isInteger(Number(e)),
    nn = (e) => e.endsWith("%") && F(e.slice(0, -1)),
    Ie = (e) => Il.test(e),
    Hl = () => !0,
    Fl = (e) => Al.test(e) && !Dl.test(e),
    Bo = () => !1,
    Bl = (e) => Ll.test(e),
    zl = (e) => Ol.test(e),
    ql = (e) => !k(e) && !S(e),
    $l = (e) => st(e, $o, Bo),
    k = (e) => Ho.test(e),
    We = (e) => st(e, Vo, Fl),
    on = (e) => st(e, Wl, F),
    Ao = (e) => st(e, zo, Bo),
    Vl = (e) => st(e, qo, zl),
    ar = (e) => st(e, Go, Bl),
    S = (e) => Fo.test(e),
    Nt = (e) => it(e, Vo),
    Gl = (e) => it(e, Yl),
    Do = (e) => it(e, zo),
    Ul = (e) => it(e, $o),
    jl = (e) => it(e, qo),
    sr = (e) => it(e, Go, !0),
    st = (e, t, r) => {
      let n = Ho.exec(e);
      return n ? (n[1] ? t(n[1]) : r(n[2])) : !1;
    },
    it = (e, t, r = !1) => {
      let n = Fo.exec(e);
      return n ? (n[1] ? t(n[1]) : r) : !1;
    },
    zo = (e) => e === "position" || e === "percentage",
    qo = (e) => e === "image" || e === "url",
    $o = (e) => e === "length" || e === "size" || e === "bg-size",
    Vo = (e) => e === "length",
    Wl = (e) => e === "number",
    Yl = (e) => e === "family-name",
    Go = (e) => e === "shadow";
  var Xl = () => {
    let e = re("color"),
      t = re("font"),
      r = re("text"),
      n = re("font-weight"),
      a = re("tracking"),
      s = re("leading"),
      i = re("breakpoint"),
      l = re("container"),
      u = re("spacing"),
      f = re("radius"),
      m = re("shadow"),
      c = re("inset-shadow"),
      d = re("text-shadow"),
      p = re("drop-shadow"),
      x = re("blur"),
      b = re("perspective"),
      v = re("aspect"),
      y = re("ease"),
      N = re("animate"),
      R = () => [
        "auto",
        "avoid",
        "all",
        "avoid-page",
        "page",
        "left",
        "right",
        "column",
      ],
      M = () => [
        "center",
        "top",
        "bottom",
        "left",
        "right",
        "top-left",
        "left-top",
        "top-right",
        "right-top",
        "bottom-right",
        "right-bottom",
        "bottom-left",
        "left-bottom",
      ],
      w = () => [...M(), S, k],
      _ = () => ["auto", "hidden", "clip", "visible", "scroll"],
      T = () => ["auto", "contain", "none"],
      C = () => [S, k, u],
      V = () => [at, "full", "auto", ...C()],
      U = () => [Be, "none", "subgrid", S, k],
      B = () => ["auto", { span: ["full", Be, S, k] }, Be, S, k],
      oe = () => [Be, "auto", S, k],
      rt = () => ["auto", "min", "max", "fr", S, k],
      vt = () => [
        "start",
        "end",
        "center",
        "between",
        "around",
        "evenly",
        "stretch",
        "baseline",
        "center-safe",
        "end-safe",
      ],
      z = () => [
        "start",
        "end",
        "center",
        "stretch",
        "center-safe",
        "end-safe",
      ],
      O = () => ["auto", ...C()],
      Q = () => [
        at,
        "auto",
        "full",
        "dvw",
        "dvh",
        "lvw",
        "lvh",
        "svw",
        "svh",
        "min",
        "max",
        "fit",
        ...C(),
      ],
      E = () => [e, S, k],
      me = () => [...M(), Do, Ao, { position: [S, k] }],
      xt = () => ["no-repeat", { repeat: ["", "x", "y", "space", "round"] }],
      Jn = () => ["auto", "cover", "contain", Ul, $l, { size: [S, k] }],
      $r = () => [nn, Nt, We],
      le = () => ["", "none", "full", f, S, k],
      pe = () => ["", F, Nt, We],
      Wt = () => ["solid", "dashed", "dotted", "double"],
      Zn = () => [
        "normal",
        "multiply",
        "screen",
        "overlay",
        "darken",
        "lighten",
        "color-dodge",
        "color-burn",
        "hard-light",
        "soft-light",
        "difference",
        "exclusion",
        "hue",
        "saturation",
        "color",
        "luminosity",
      ],
      ae = () => [F, nn, Do, Ao],
      eo = () => ["", "none", x, S, k],
      Yt = () => ["none", F, S, k],
      Xt = () => ["none", F, S, k],
      Vr = () => [F, S, k],
      Qt = () => [at, "full", ...C()];
    return {
      cacheSize: 500,
      theme: {
        animate: ["spin", "ping", "pulse", "bounce"],
        aspect: ["video"],
        blur: [Ie],
        breakpoint: [Ie],
        color: [Hl],
        container: [Ie],
        "drop-shadow": [Ie],
        ease: ["in", "out", "in-out"],
        font: [ql],
        "font-weight": [
          "thin",
          "extralight",
          "light",
          "normal",
          "medium",
          "semibold",
          "bold",
          "extrabold",
          "black",
        ],
        "inset-shadow": [Ie],
        leading: ["none", "tight", "snug", "normal", "relaxed", "loose"],
        perspective: [
          "dramatic",
          "near",
          "normal",
          "midrange",
          "distant",
          "none",
        ],
        radius: [Ie],
        shadow: [Ie],
        spacing: ["px", F],
        text: [Ie],
        "text-shadow": [Ie],
        tracking: ["tighter", "tight", "normal", "wide", "wider", "widest"],
      },
      classGroups: {
        aspect: [{ aspect: ["auto", "square", at, k, S, v] }],
        container: ["container"],
        columns: [{ columns: [F, k, S, l] }],
        "break-after": [{ "break-after": R() }],
        "break-before": [{ "break-before": R() }],
        "break-inside": [
          { "break-inside": ["auto", "avoid", "avoid-page", "avoid-column"] },
        ],
        "box-decoration": [{ "box-decoration": ["slice", "clone"] }],
        box: [{ box: ["border", "content"] }],
        display: [
          "block",
          "inline-block",
          "inline",
          "flex",
          "inline-flex",
          "table",
          "inline-table",
          "table-caption",
          "table-cell",
          "table-column",
          "table-column-group",
          "table-footer-group",
          "table-header-group",
          "table-row-group",
          "table-row",
          "flow-root",
          "grid",
          "inline-grid",
          "contents",
          "list-item",
          "hidden",
        ],
        sr: ["sr-only", "not-sr-only"],
        float: [{ float: ["right", "left", "none", "start", "end"] }],
        clear: [{ clear: ["left", "right", "both", "none", "start", "end"] }],
        isolation: ["isolate", "isolation-auto"],
        "object-fit": [
          { object: ["contain", "cover", "fill", "none", "scale-down"] },
        ],
        "object-position": [{ object: w() }],
        overflow: [{ overflow: _() }],
        "overflow-x": [{ "overflow-x": _() }],
        "overflow-y": [{ "overflow-y": _() }],
        overscroll: [{ overscroll: T() }],
        "overscroll-x": [{ "overscroll-x": T() }],
        "overscroll-y": [{ "overscroll-y": T() }],
        position: ["static", "fixed", "absolute", "relative", "sticky"],
        inset: [{ inset: V() }],
        "inset-x": [{ "inset-x": V() }],
        "inset-y": [{ "inset-y": V() }],
        start: [{ start: V() }],
        end: [{ end: V() }],
        top: [{ top: V() }],
        right: [{ right: V() }],
        bottom: [{ bottom: V() }],
        left: [{ left: V() }],
        visibility: ["visible", "invisible", "collapse"],
        z: [{ z: [Be, "auto", S, k] }],
        basis: [{ basis: [at, "full", "auto", l, ...C()] }],
        "flex-direction": [
          { flex: ["row", "row-reverse", "col", "col-reverse"] },
        ],
        "flex-wrap": [{ flex: ["nowrap", "wrap", "wrap-reverse"] }],
        flex: [{ flex: [F, at, "auto", "initial", "none", k] }],
        grow: [{ grow: ["", F, S, k] }],
        shrink: [{ shrink: ["", F, S, k] }],
        order: [{ order: [Be, "first", "last", "none", S, k] }],
        "grid-cols": [{ "grid-cols": U() }],
        "col-start-end": [{ col: B() }],
        "col-start": [{ "col-start": oe() }],
        "col-end": [{ "col-end": oe() }],
        "grid-rows": [{ "grid-rows": U() }],
        "row-start-end": [{ row: B() }],
        "row-start": [{ "row-start": oe() }],
        "row-end": [{ "row-end": oe() }],
        "grid-flow": [
          { "grid-flow": ["row", "col", "dense", "row-dense", "col-dense"] },
        ],
        "auto-cols": [{ "auto-cols": rt() }],
        "auto-rows": [{ "auto-rows": rt() }],
        gap: [{ gap: C() }],
        "gap-x": [{ "gap-x": C() }],
        "gap-y": [{ "gap-y": C() }],
        "justify-content": [{ justify: [...vt(), "normal"] }],
        "justify-items": [{ "justify-items": [...z(), "normal"] }],
        "justify-self": [{ "justify-self": ["auto", ...z()] }],
        "align-content": [{ content: ["normal", ...vt()] }],
        "align-items": [{ items: [...z(), { baseline: ["", "last"] }] }],
        "align-self": [{ self: ["auto", ...z(), { baseline: ["", "last"] }] }],
        "place-content": [{ "place-content": vt() }],
        "place-items": [{ "place-items": [...z(), "baseline"] }],
        "place-self": [{ "place-self": ["auto", ...z()] }],
        p: [{ p: C() }],
        px: [{ px: C() }],
        py: [{ py: C() }],
        ps: [{ ps: C() }],
        pe: [{ pe: C() }],
        pt: [{ pt: C() }],
        pr: [{ pr: C() }],
        pb: [{ pb: C() }],
        pl: [{ pl: C() }],
        m: [{ m: O() }],
        mx: [{ mx: O() }],
        my: [{ my: O() }],
        ms: [{ ms: O() }],
        me: [{ me: O() }],
        mt: [{ mt: O() }],
        mr: [{ mr: O() }],
        mb: [{ mb: O() }],
        ml: [{ ml: O() }],
        "space-x": [{ "space-x": C() }],
        "space-x-reverse": ["space-x-reverse"],
        "space-y": [{ "space-y": C() }],
        "space-y-reverse": ["space-y-reverse"],
        size: [{ size: Q() }],
        w: [{ w: [l, "screen", ...Q()] }],
        "min-w": [{ "min-w": [l, "screen", "none", ...Q()] }],
        "max-w": [
          { "max-w": [l, "screen", "none", "prose", { screen: [i] }, ...Q()] },
        ],
        h: [{ h: ["screen", "lh", ...Q()] }],
        "min-h": [{ "min-h": ["screen", "lh", "none", ...Q()] }],
        "max-h": [{ "max-h": ["screen", "lh", ...Q()] }],
        "font-size": [{ text: ["base", r, Nt, We] }],
        "font-smoothing": ["antialiased", "subpixel-antialiased"],
        "font-style": ["italic", "not-italic"],
        "font-weight": [{ font: [n, S, on] }],
        "font-stretch": [
          {
            "font-stretch": [
              "ultra-condensed",
              "extra-condensed",
              "condensed",
              "semi-condensed",
              "normal",
              "semi-expanded",
              "expanded",
              "extra-expanded",
              "ultra-expanded",
              nn,
              k,
            ],
          },
        ],
        "font-family": [{ font: [Gl, k, t] }],
        "fvn-normal": ["normal-nums"],
        "fvn-ordinal": ["ordinal"],
        "fvn-slashed-zero": ["slashed-zero"],
        "fvn-figure": ["lining-nums", "oldstyle-nums"],
        "fvn-spacing": ["proportional-nums", "tabular-nums"],
        "fvn-fraction": ["diagonal-fractions", "stacked-fractions"],
        tracking: [{ tracking: [a, S, k] }],
        "line-clamp": [{ "line-clamp": [F, "none", S, on] }],
        leading: [{ leading: [s, ...C()] }],
        "list-image": [{ "list-image": ["none", S, k] }],
        "list-style-position": [{ list: ["inside", "outside"] }],
        "list-style-type": [{ list: ["disc", "decimal", "none", S, k] }],
        "text-alignment": [
          { text: ["left", "center", "right", "justify", "start", "end"] },
        ],
        "placeholder-color": [{ placeholder: E() }],
        "text-color": [{ text: E() }],
        "text-decoration": [
          "underline",
          "overline",
          "line-through",
          "no-underline",
        ],
        "text-decoration-style": [{ decoration: [...Wt(), "wavy"] }],
        "text-decoration-thickness": [
          { decoration: [F, "from-font", "auto", S, We] },
        ],
        "text-decoration-color": [{ decoration: E() }],
        "underline-offset": [{ "underline-offset": [F, "auto", S, k] }],
        "text-transform": [
          "uppercase",
          "lowercase",
          "capitalize",
          "normal-case",
        ],
        "text-overflow": ["truncate", "text-ellipsis", "text-clip"],
        "text-wrap": [{ text: ["wrap", "nowrap", "balance", "pretty"] }],
        indent: [{ indent: C() }],
        "vertical-align": [
          {
            align: [
              "baseline",
              "top",
              "middle",
              "bottom",
              "text-top",
              "text-bottom",
              "sub",
              "super",
              S,
              k,
            ],
          },
        ],
        whitespace: [
          {
            whitespace: [
              "normal",
              "nowrap",
              "pre",
              "pre-line",
              "pre-wrap",
              "break-spaces",
            ],
          },
        ],
        break: [{ break: ["normal", "words", "all", "keep"] }],
        wrap: [{ wrap: ["break-word", "anywhere", "normal"] }],
        hyphens: [{ hyphens: ["none", "manual", "auto"] }],
        content: [{ content: ["none", S, k] }],
        "bg-attachment": [{ bg: ["fixed", "local", "scroll"] }],
        "bg-clip": [{ "bg-clip": ["border", "padding", "content", "text"] }],
        "bg-origin": [{ "bg-origin": ["border", "padding", "content"] }],
        "bg-position": [{ bg: me() }],
        "bg-repeat": [{ bg: xt() }],
        "bg-size": [{ bg: Jn() }],
        "bg-image": [
          {
            bg: [
              "none",
              {
                linear: [
                  { to: ["t", "tr", "r", "br", "b", "bl", "l", "tl"] },
                  Be,
                  S,
                  k,
                ],
                radial: ["", S, k],
                conic: [Be, S, k],
              },
              jl,
              Vl,
            ],
          },
        ],
        "bg-color": [{ bg: E() }],
        "gradient-from-pos": [{ from: $r() }],
        "gradient-via-pos": [{ via: $r() }],
        "gradient-to-pos": [{ to: $r() }],
        "gradient-from": [{ from: E() }],
        "gradient-via": [{ via: E() }],
        "gradient-to": [{ to: E() }],
        rounded: [{ rounded: le() }],
        "rounded-s": [{ "rounded-s": le() }],
        "rounded-e": [{ "rounded-e": le() }],
        "rounded-t": [{ "rounded-t": le() }],
        "rounded-r": [{ "rounded-r": le() }],
        "rounded-b": [{ "rounded-b": le() }],
        "rounded-l": [{ "rounded-l": le() }],
        "rounded-ss": [{ "rounded-ss": le() }],
        "rounded-se": [{ "rounded-se": le() }],
        "rounded-ee": [{ "rounded-ee": le() }],
        "rounded-es": [{ "rounded-es": le() }],
        "rounded-tl": [{ "rounded-tl": le() }],
        "rounded-tr": [{ "rounded-tr": le() }],
        "rounded-br": [{ "rounded-br": le() }],
        "rounded-bl": [{ "rounded-bl": le() }],
        "border-w": [{ border: pe() }],
        "border-w-x": [{ "border-x": pe() }],
        "border-w-y": [{ "border-y": pe() }],
        "border-w-s": [{ "border-s": pe() }],
        "border-w-e": [{ "border-e": pe() }],
        "border-w-t": [{ "border-t": pe() }],
        "border-w-r": [{ "border-r": pe() }],
        "border-w-b": [{ "border-b": pe() }],
        "border-w-l": [{ "border-l": pe() }],
        "divide-x": [{ "divide-x": pe() }],
        "divide-x-reverse": ["divide-x-reverse"],
        "divide-y": [{ "divide-y": pe() }],
        "divide-y-reverse": ["divide-y-reverse"],
        "border-style": [{ border: [...Wt(), "hidden", "none"] }],
        "divide-style": [{ divide: [...Wt(), "hidden", "none"] }],
        "border-color": [{ border: E() }],
        "border-color-x": [{ "border-x": E() }],
        "border-color-y": [{ "border-y": E() }],
        "border-color-s": [{ "border-s": E() }],
        "border-color-e": [{ "border-e": E() }],
        "border-color-t": [{ "border-t": E() }],
        "border-color-r": [{ "border-r": E() }],
        "border-color-b": [{ "border-b": E() }],
        "border-color-l": [{ "border-l": E() }],
        "divide-color": [{ divide: E() }],
        "outline-style": [{ outline: [...Wt(), "none", "hidden"] }],
        "outline-offset": [{ "outline-offset": [F, S, k] }],
        "outline-w": [{ outline: ["", F, Nt, We] }],
        "outline-color": [{ outline: E() }],
        shadow: [{ shadow: ["", "none", m, sr, ar] }],
        "shadow-color": [{ shadow: E() }],
        "inset-shadow": [{ "inset-shadow": ["none", c, sr, ar] }],
        "inset-shadow-color": [{ "inset-shadow": E() }],
        "ring-w": [{ ring: pe() }],
        "ring-w-inset": ["ring-inset"],
        "ring-color": [{ ring: E() }],
        "ring-offset-w": [{ "ring-offset": [F, We] }],
        "ring-offset-color": [{ "ring-offset": E() }],
        "inset-ring-w": [{ "inset-ring": pe() }],
        "inset-ring-color": [{ "inset-ring": E() }],
        "text-shadow": [{ "text-shadow": ["none", d, sr, ar] }],
        "text-shadow-color": [{ "text-shadow": E() }],
        opacity: [{ opacity: [F, S, k] }],
        "mix-blend": [
          { "mix-blend": [...Zn(), "plus-darker", "plus-lighter"] },
        ],
        "bg-blend": [{ "bg-blend": Zn() }],
        "mask-clip": [
          {
            "mask-clip": [
              "border",
              "padding",
              "content",
              "fill",
              "stroke",
              "view",
            ],
          },
          "mask-no-clip",
        ],
        "mask-composite": [
          { mask: ["add", "subtract", "intersect", "exclude"] },
        ],
        "mask-image-linear-pos": [{ "mask-linear": [F] }],
        "mask-image-linear-from-pos": [{ "mask-linear-from": ae() }],
        "mask-image-linear-to-pos": [{ "mask-linear-to": ae() }],
        "mask-image-linear-from-color": [{ "mask-linear-from": E() }],
        "mask-image-linear-to-color": [{ "mask-linear-to": E() }],
        "mask-image-t-from-pos": [{ "mask-t-from": ae() }],
        "mask-image-t-to-pos": [{ "mask-t-to": ae() }],
        "mask-image-t-from-color": [{ "mask-t-from": E() }],
        "mask-image-t-to-color": [{ "mask-t-to": E() }],
        "mask-image-r-from-pos": [{ "mask-r-from": ae() }],
        "mask-image-r-to-pos": [{ "mask-r-to": ae() }],
        "mask-image-r-from-color": [{ "mask-r-from": E() }],
        "mask-image-r-to-color": [{ "mask-r-to": E() }],
        "mask-image-b-from-pos": [{ "mask-b-from": ae() }],
        "mask-image-b-to-pos": [{ "mask-b-to": ae() }],
        "mask-image-b-from-color": [{ "mask-b-from": E() }],
        "mask-image-b-to-color": [{ "mask-b-to": E() }],
        "mask-image-l-from-pos": [{ "mask-l-from": ae() }],
        "mask-image-l-to-pos": [{ "mask-l-to": ae() }],
        "mask-image-l-from-color": [{ "mask-l-from": E() }],
        "mask-image-l-to-color": [{ "mask-l-to": E() }],
        "mask-image-x-from-pos": [{ "mask-x-from": ae() }],
        "mask-image-x-to-pos": [{ "mask-x-to": ae() }],
        "mask-image-x-from-color": [{ "mask-x-from": E() }],
        "mask-image-x-to-color": [{ "mask-x-to": E() }],
        "mask-image-y-from-pos": [{ "mask-y-from": ae() }],
        "mask-image-y-to-pos": [{ "mask-y-to": ae() }],
        "mask-image-y-from-color": [{ "mask-y-from": E() }],
        "mask-image-y-to-color": [{ "mask-y-to": E() }],
        "mask-image-radial": [{ "mask-radial": [S, k] }],
        "mask-image-radial-from-pos": [{ "mask-radial-from": ae() }],
        "mask-image-radial-to-pos": [{ "mask-radial-to": ae() }],
        "mask-image-radial-from-color": [{ "mask-radial-from": E() }],
        "mask-image-radial-to-color": [{ "mask-radial-to": E() }],
        "mask-image-radial-shape": [{ "mask-radial": ["circle", "ellipse"] }],
        "mask-image-radial-size": [
          {
            "mask-radial": [
              { closest: ["side", "corner"], farthest: ["side", "corner"] },
            ],
          },
        ],
        "mask-image-radial-pos": [{ "mask-radial-at": M() }],
        "mask-image-conic-pos": [{ "mask-conic": [F] }],
        "mask-image-conic-from-pos": [{ "mask-conic-from": ae() }],
        "mask-image-conic-to-pos": [{ "mask-conic-to": ae() }],
        "mask-image-conic-from-color": [{ "mask-conic-from": E() }],
        "mask-image-conic-to-color": [{ "mask-conic-to": E() }],
        "mask-mode": [{ mask: ["alpha", "luminance", "match"] }],
        "mask-origin": [
          {
            "mask-origin": [
              "border",
              "padding",
              "content",
              "fill",
              "stroke",
              "view",
            ],
          },
        ],
        "mask-position": [{ mask: me() }],
        "mask-repeat": [{ mask: xt() }],
        "mask-size": [{ mask: Jn() }],
        "mask-type": [{ "mask-type": ["alpha", "luminance"] }],
        "mask-image": [{ mask: ["none", S, k] }],
        filter: [{ filter: ["", "none", S, k] }],
        blur: [{ blur: eo() }],
        brightness: [{ brightness: [F, S, k] }],
        contrast: [{ contrast: [F, S, k] }],
        "drop-shadow": [{ "drop-shadow": ["", "none", p, sr, ar] }],
        "drop-shadow-color": [{ "drop-shadow": E() }],
        grayscale: [{ grayscale: ["", F, S, k] }],
        "hue-rotate": [{ "hue-rotate": [F, S, k] }],
        invert: [{ invert: ["", F, S, k] }],
        saturate: [{ saturate: [F, S, k] }],
        sepia: [{ sepia: ["", F, S, k] }],
        "backdrop-filter": [{ "backdrop-filter": ["", "none", S, k] }],
        "backdrop-blur": [{ "backdrop-blur": eo() }],
        "backdrop-brightness": [{ "backdrop-brightness": [F, S, k] }],
        "backdrop-contrast": [{ "backdrop-contrast": [F, S, k] }],
        "backdrop-grayscale": [{ "backdrop-grayscale": ["", F, S, k] }],
        "backdrop-hue-rotate": [{ "backdrop-hue-rotate": [F, S, k] }],
        "backdrop-invert": [{ "backdrop-invert": ["", F, S, k] }],
        "backdrop-opacity": [{ "backdrop-opacity": [F, S, k] }],
        "backdrop-saturate": [{ "backdrop-saturate": [F, S, k] }],
        "backdrop-sepia": [{ "backdrop-sepia": ["", F, S, k] }],
        "border-collapse": [{ border: ["collapse", "separate"] }],
        "border-spacing": [{ "border-spacing": C() }],
        "border-spacing-x": [{ "border-spacing-x": C() }],
        "border-spacing-y": [{ "border-spacing-y": C() }],
        "table-layout": [{ table: ["auto", "fixed"] }],
        caption: [{ caption: ["top", "bottom"] }],
        transition: [
          {
            transition: [
              "",
              "all",
              "colors",
              "opacity",
              "shadow",
              "transform",
              "none",
              S,
              k,
            ],
          },
        ],
        "transition-behavior": [{ transition: ["normal", "discrete"] }],
        duration: [{ duration: [F, "initial", S, k] }],
        ease: [{ ease: ["linear", "initial", y, S, k] }],
        delay: [{ delay: [F, S, k] }],
        animate: [{ animate: ["none", N, S, k] }],
        backface: [{ backface: ["hidden", "visible"] }],
        perspective: [{ perspective: [b, S, k] }],
        "perspective-origin": [{ "perspective-origin": w() }],
        rotate: [{ rotate: Yt() }],
        "rotate-x": [{ "rotate-x": Yt() }],
        "rotate-y": [{ "rotate-y": Yt() }],
        "rotate-z": [{ "rotate-z": Yt() }],
        scale: [{ scale: Xt() }],
        "scale-x": [{ "scale-x": Xt() }],
        "scale-y": [{ "scale-y": Xt() }],
        "scale-z": [{ "scale-z": Xt() }],
        "scale-3d": ["scale-3d"],
        skew: [{ skew: Vr() }],
        "skew-x": [{ "skew-x": Vr() }],
        "skew-y": [{ "skew-y": Vr() }],
        transform: [{ transform: [S, k, "", "none", "gpu", "cpu"] }],
        "transform-origin": [{ origin: w() }],
        "transform-style": [{ transform: ["3d", "flat"] }],
        translate: [{ translate: Qt() }],
        "translate-x": [{ "translate-x": Qt() }],
        "translate-y": [{ "translate-y": Qt() }],
        "translate-z": [{ "translate-z": Qt() }],
        "translate-none": ["translate-none"],
        accent: [{ accent: E() }],
        appearance: [{ appearance: ["none", "auto"] }],
        "caret-color": [{ caret: E() }],
        "color-scheme": [
          {
            scheme: [
              "normal",
              "dark",
              "light",
              "light-dark",
              "only-dark",
              "only-light",
            ],
          },
        ],
        cursor: [
          {
            cursor: [
              "auto",
              "default",
              "pointer",
              "wait",
              "text",
              "move",
              "help",
              "not-allowed",
              "none",
              "context-menu",
              "progress",
              "cell",
              "crosshair",
              "vertical-text",
              "alias",
              "copy",
              "no-drop",
              "grab",
              "grabbing",
              "all-scroll",
              "col-resize",
              "row-resize",
              "n-resize",
              "e-resize",
              "s-resize",
              "w-resize",
              "ne-resize",
              "nw-resize",
              "se-resize",
              "sw-resize",
              "ew-resize",
              "ns-resize",
              "nesw-resize",
              "nwse-resize",
              "zoom-in",
              "zoom-out",
              S,
              k,
            ],
          },
        ],
        "field-sizing": [{ "field-sizing": ["fixed", "content"] }],
        "pointer-events": [{ "pointer-events": ["auto", "none"] }],
        resize: [{ resize: ["none", "", "y", "x"] }],
        "scroll-behavior": [{ scroll: ["auto", "smooth"] }],
        "scroll-m": [{ "scroll-m": C() }],
        "scroll-mx": [{ "scroll-mx": C() }],
        "scroll-my": [{ "scroll-my": C() }],
        "scroll-ms": [{ "scroll-ms": C() }],
        "scroll-me": [{ "scroll-me": C() }],
        "scroll-mt": [{ "scroll-mt": C() }],
        "scroll-mr": [{ "scroll-mr": C() }],
        "scroll-mb": [{ "scroll-mb": C() }],
        "scroll-ml": [{ "scroll-ml": C() }],
        "scroll-p": [{ "scroll-p": C() }],
        "scroll-px": [{ "scroll-px": C() }],
        "scroll-py": [{ "scroll-py": C() }],
        "scroll-ps": [{ "scroll-ps": C() }],
        "scroll-pe": [{ "scroll-pe": C() }],
        "scroll-pt": [{ "scroll-pt": C() }],
        "scroll-pr": [{ "scroll-pr": C() }],
        "scroll-pb": [{ "scroll-pb": C() }],
        "scroll-pl": [{ "scroll-pl": C() }],
        "snap-align": [{ snap: ["start", "end", "center", "align-none"] }],
        "snap-stop": [{ snap: ["normal", "always"] }],
        "snap-type": [{ snap: ["none", "x", "y", "both"] }],
        "snap-strictness": [{ snap: ["mandatory", "proximity"] }],
        touch: [{ touch: ["auto", "none", "manipulation"] }],
        "touch-x": [{ "touch-pan": ["x", "left", "right"] }],
        "touch-y": [{ "touch-pan": ["y", "up", "down"] }],
        "touch-pz": ["touch-pinch-zoom"],
        select: [{ select: ["none", "text", "all", "auto"] }],
        "will-change": [
          { "will-change": ["auto", "scroll", "contents", "transform", S, k] },
        ],
        fill: [{ fill: ["none", ...E()] }],
        "stroke-w": [{ stroke: [F, Nt, We, on] }],
        stroke: [{ stroke: ["none", ...E()] }],
        "forced-color-adjust": [{ "forced-color-adjust": ["auto", "none"] }],
      },
      conflictingClassGroups: {
        overflow: ["overflow-x", "overflow-y"],
        overscroll: ["overscroll-x", "overscroll-y"],
        inset: [
          "inset-x",
          "inset-y",
          "start",
          "end",
          "top",
          "right",
          "bottom",
          "left",
        ],
        "inset-x": ["right", "left"],
        "inset-y": ["top", "bottom"],
        flex: ["basis", "grow", "shrink"],
        gap: ["gap-x", "gap-y"],
        p: ["px", "py", "ps", "pe", "pt", "pr", "pb", "pl"],
        px: ["pr", "pl"],
        py: ["pt", "pb"],
        m: ["mx", "my", "ms", "me", "mt", "mr", "mb", "ml"],
        mx: ["mr", "ml"],
        my: ["mt", "mb"],
        size: ["w", "h"],
        "font-size": ["leading"],
        "fvn-normal": [
          "fvn-ordinal",
          "fvn-slashed-zero",
          "fvn-figure",
          "fvn-spacing",
          "fvn-fraction",
        ],
        "fvn-ordinal": ["fvn-normal"],
        "fvn-slashed-zero": ["fvn-normal"],
        "fvn-figure": ["fvn-normal"],
        "fvn-spacing": ["fvn-normal"],
        "fvn-fraction": ["fvn-normal"],
        "line-clamp": ["display", "overflow"],
        rounded: [
          "rounded-s",
          "rounded-e",
          "rounded-t",
          "rounded-r",
          "rounded-b",
          "rounded-l",
          "rounded-ss",
          "rounded-se",
          "rounded-ee",
          "rounded-es",
          "rounded-tl",
          "rounded-tr",
          "rounded-br",
          "rounded-bl",
        ],
        "rounded-s": ["rounded-ss", "rounded-es"],
        "rounded-e": ["rounded-se", "rounded-ee"],
        "rounded-t": ["rounded-tl", "rounded-tr"],
        "rounded-r": ["rounded-tr", "rounded-br"],
        "rounded-b": ["rounded-br", "rounded-bl"],
        "rounded-l": ["rounded-tl", "rounded-bl"],
        "border-spacing": ["border-spacing-x", "border-spacing-y"],
        "border-w": [
          "border-w-x",
          "border-w-y",
          "border-w-s",
          "border-w-e",
          "border-w-t",
          "border-w-r",
          "border-w-b",
          "border-w-l",
        ],
        "border-w-x": ["border-w-r", "border-w-l"],
        "border-w-y": ["border-w-t", "border-w-b"],
        "border-color": [
          "border-color-x",
          "border-color-y",
          "border-color-s",
          "border-color-e",
          "border-color-t",
          "border-color-r",
          "border-color-b",
          "border-color-l",
        ],
        "border-color-x": ["border-color-r", "border-color-l"],
        "border-color-y": ["border-color-t", "border-color-b"],
        translate: ["translate-x", "translate-y", "translate-none"],
        "translate-none": [
          "translate",
          "translate-x",
          "translate-y",
          "translate-z",
        ],
        "scroll-m": [
          "scroll-mx",
          "scroll-my",
          "scroll-ms",
          "scroll-me",
          "scroll-mt",
          "scroll-mr",
          "scroll-mb",
          "scroll-ml",
        ],
        "scroll-mx": ["scroll-mr", "scroll-ml"],
        "scroll-my": ["scroll-mt", "scroll-mb"],
        "scroll-p": [
          "scroll-px",
          "scroll-py",
          "scroll-ps",
          "scroll-pe",
          "scroll-pt",
          "scroll-pr",
          "scroll-pb",
          "scroll-pl",
        ],
        "scroll-px": ["scroll-pr", "scroll-pl"],
        "scroll-py": ["scroll-pt", "scroll-pb"],
        touch: ["touch-x", "touch-y", "touch-pz"],
        "touch-x": ["touch"],
        "touch-y": ["touch"],
        "touch-pz": ["touch"],
      },
      conflictingClassGroupModifiers: { "font-size": ["leading"] },
      orderSensitiveModifiers: [
        "*",
        "**",
        "after",
        "backdrop",
        "before",
        "details-content",
        "file",
        "first-letter",
        "first-line",
        "marker",
        "placeholder",
        "selection",
      ],
    };
  };
  var Uo = Ml(Xl);
  function D(...e) {
    return Uo(or(e));
  }
  var ee = {};
  il(ee, {
    Children: () => Te,
    Component: () => he,
    Fragment: () => h,
    PureComponent: () => ir,
    StrictMode: () => va,
    Suspense: () => Ct,
    SuspenseList: () => lt,
    __SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED: () => ua,
    cloneElement: () => Ye,
    createContext: () => Ee,
    createElement: () => g,
    createFactory: () => fa,
    createPortal: () => sa,
    createRef: () => Zt,
    default: () => be,
    findDOMNode: () => ha,
    flushSync: () => xn,
    forwardRef: () => H,
    hydrate: () => ca,
    isElement: () => xa,
    isFragment: () => da,
    isMemo: () => ma,
    isValidElement: () => Re,
    lazy: () => aa,
    memo: () => ta,
    render: () => la,
    startTransition: () => mn,
    unmountComponentAtNode: () => pa,
    unstable_batchedUpdates: () => ga,
    useCallback: () => ce,
    useContext: () => Fe,
    useDebugValue: () => rr,
    useDeferredValue: () => pn,
    useEffect: () => q,
    useErrorBoundary: () => pl,
    useId: () => nr,
    useImperativeHandle: () => tr,
    useInsertionEffect: () => gn,
    useLayoutEffect: () => Pe,
    useMemo: () => Z,
    useReducer: () => He,
    useRef: () => $,
    useState: () => P,
    useSyncExternalStore: () => dn,
    useTransition: () => hn,
    version: () => cc,
  });
  function ea(e, t) {
    for (var r in t) e[r] = t[r];
    return e;
  }
  function fn(e, t) {
    for (var r in e) if (r !== "__source" && !(r in t)) return !0;
    for (var n in t) if (n !== "__source" && e[n] !== t[n]) return !0;
    return !1;
  }
  function dn(e, t) {
    var r = t(),
      n = P({ t: { __: r, u: t } }),
      a = n[0].t,
      s = n[1];
    return (
      Pe(
        function () {
          (a.__ = r), (a.u = t), un(a) && s({ t: a });
        },
        [e, r, t]
      ),
      q(
        function () {
          return (
            un(a) && s({ t: a }),
            e(function () {
              un(a) && s({ t: a });
            })
          );
        },
        [e]
      ),
      r
    );
  }
  function un(e) {
    var t,
      r,
      n = e.u,
      a = e.__;
    try {
      var s = n();
      return !(
        ((t = a) === (r = s) && (t !== 0 || 1 / t == 1 / r)) ||
        (t != t && r != r)
      );
    } catch {
      return !0;
    }
  }
  function mn(e) {
    e();
  }
  function pn(e) {
    return e;
  }
  function hn() {
    return [!1, mn];
  }
  var gn = Pe;
  function ir(e, t) {
    (this.props = e), (this.context = t);
  }
  function ta(e, t) {
    function r(a) {
      var s = this.props.ref,
        i = s == a.ref;
      return (
        !i && s && (s.call ? s(null) : (s.current = null)),
        t ? !t(this.props, a) || !i : fn(this.props, a)
      );
    }
    function n(a) {
      return (this.shouldComponentUpdate = r), g(e, a);
    }
    return (
      (n.displayName = "Memo(" + (e.displayName || e.name) + ")"),
      (n.prototype.isReactComponent = !0),
      (n.__f = !0),
      (n.type = e),
      n
    );
  }
  ((ir.prototype = new he()).isPureReactComponent = !0),
    (ir.prototype.shouldComponentUpdate = function (e, t) {
      return fn(this.props, e) || fn(this.state, t);
    });
  var jo = L.__b;
  L.__b = function (e) {
    e.type && e.type.__f && e.ref && ((e.props.ref = e.ref), (e.ref = null)),
      jo && jo(e);
  };
  var Ql =
    (typeof Symbol < "u" && Symbol.for && Symbol.for("react.forward_ref")) ||
    3911;
  function H(e) {
    function t(r) {
      var n = ea({}, r);
      return delete n.ref, e(n, r.ref || null);
    }
    return (
      (t.$$typeof = Ql),
      (t.render = e),
      (t.prototype.isReactComponent = t.__f = !0),
      (t.displayName = "ForwardRef(" + (e.displayName || e.name) + ")"),
      t
    );
  }
  var Wo = function (e, t) {
      return e == null ? null : Ce(Ce(e).map(t));
    },
    Te = {
      map: Wo,
      forEach: Wo,
      count: function (e) {
        return e ? Ce(e).length : 0;
      },
      only: function (e) {
        var t = Ce(e);
        if (t.length !== 1) throw "Children.only";
        return t[0];
      },
      toArray: Ce,
    },
    Kl = L.__e;
  L.__e = function (e, t, r, n) {
    if (e.then) {
      for (var a, s = t; (s = s.__); )
        if ((a = s.__c) && a.__c)
          return (
            t.__e == null && ((t.__e = r.__e), (t.__k = r.__k)), a.__c(e, t)
          );
    }
    Kl(e, t, r, n);
  };
  var Yo = L.unmount;
  function ra(e, t, r) {
    return (
      e &&
        (e.__c &&
          e.__c.__H &&
          (e.__c.__H.__.forEach(function (n) {
            typeof n.__c == "function" && n.__c();
          }),
          (e.__c.__H = null)),
        (e = ea({}, e)).__c != null &&
          (e.__c.__P === r && (e.__c.__P = t),
          (e.__c.__e = !0),
          (e.__c = null)),
        (e.__k =
          e.__k &&
          e.__k.map(function (n) {
            return ra(n, t, r);
          }))),
      e
    );
  }
  function na(e, t, r) {
    return (
      e &&
        r &&
        ((e.__v = null),
        (e.__k =
          e.__k &&
          e.__k.map(function (n) {
            return na(n, t, r);
          })),
        e.__c &&
          e.__c.__P === t &&
          (e.__e && r.appendChild(e.__e), (e.__c.__e = !0), (e.__c.__P = r))),
      e
    );
  }
  function Ct() {
    (this.__u = 0), (this.o = null), (this.__b = null);
  }
  function oa(e) {
    var t = e.__.__c;
    return t && t.__a && t.__a(e);
  }
  function aa(e) {
    var t, r, n;
    function a(s) {
      if (
        (t ||
          (t = e()).then(
            function (i) {
              r = i.default || i;
            },
            function (i) {
              n = i;
            }
          ),
        n)
      )
        throw n;
      if (!r) throw t;
      return g(r, s);
    }
    return (a.displayName = "Lazy"), (a.__f = !0), a;
  }
  function lt() {
    (this.i = null), (this.l = null);
  }
  (L.unmount = function (e) {
    var t = e.__c;
    t && t.__R && t.__R(), t && 32 & e.__u && (e.type = null), Yo && Yo(e);
  }),
    ((Ct.prototype = new he()).__c = function (e, t) {
      var r = t.__c,
        n = this;
      n.o == null && (n.o = []), n.o.push(r);
      var a = oa(n.__v),
        s = !1,
        i = function () {
          s || ((s = !0), (r.__R = null), a ? a(l) : l());
        };
      r.__R = i;
      var l = function () {
        if (!--n.__u) {
          if (n.state.__a) {
            var u = n.state.__a;
            n.__v.__k[0] = na(u, u.__c.__P, u.__c.__O);
          }
          var f;
          for (n.setState({ __a: (n.__b = null) }); (f = n.o.pop()); )
            f.forceUpdate();
        }
      };
      n.__u++ || 32 & t.__u || n.setState({ __a: (n.__b = n.__v.__k[0]) }),
        e.then(i, i);
    }),
    (Ct.prototype.componentWillUnmount = function () {
      this.o = [];
    }),
    (Ct.prototype.render = function (e, t) {
      if (this.__b) {
        if (this.__v.__k) {
          var r = document.createElement("div"),
            n = this.__v.__k[0].__c;
          this.__v.__k[0] = ra(this.__b, r, (n.__O = n.__P));
        }
        this.__b = null;
      }
      var a = t.__a && g(h, null, e.fallback);
      return a && (a.__u &= -33), [g(h, null, t.__a ? null : e.children), a];
    });
  var Xo = function (e, t, r) {
    if (
      (++r[1] === r[0] && e.l.delete(t),
      e.props.revealOrder && (e.props.revealOrder[0] !== "t" || !e.l.size))
    )
      for (r = e.i; r; ) {
        for (; r.length > 3; ) r.pop()();
        if (r[1] < r[0]) break;
        e.i = r = r[2];
      }
  };
  function Jl(e) {
    return (
      (this.getChildContext = function () {
        return e.context;
      }),
      e.children
    );
  }
  function Zl(e) {
    var t = this,
      r = e.h;
    if (
      ((t.componentWillUnmount = function () {
        Oe(null, t.v), (t.v = null), (t.h = null);
      }),
      t.h && t.h !== r && t.componentWillUnmount(),
      !t.v)
    ) {
      for (var n = t.__v; n !== null && !n.__m && n.__ !== null; ) n = n.__;
      (t.h = r),
        (t.v = {
          nodeType: 1,
          parentNode: r,
          childNodes: [],
          __k: { __m: n.__m },
          contains: function () {
            return !0;
          },
          insertBefore: function (a, s) {
            this.childNodes.push(a), t.h.insertBefore(a, s);
          },
          removeChild: function (a) {
            this.childNodes.splice(this.childNodes.indexOf(a) >>> 1, 1),
              t.h.removeChild(a);
          },
        });
    }
    Oe(g(Jl, { context: t.context }, e.__v), t.v);
  }
  function sa(e, t) {
    var r = g(Zl, { __v: e, h: t });
    return (r.containerInfo = t), r;
  }
  ((lt.prototype = new he()).__a = function (e) {
    var t = this,
      r = oa(t.__v),
      n = t.l.get(e);
    return (
      n[0]++,
      function (a) {
        var s = function () {
          t.props.revealOrder ? (n.push(a), Xo(t, e, n)) : a();
        };
        r ? r(s) : s();
      }
    );
  }),
    (lt.prototype.render = function (e) {
      (this.i = null), (this.l = new Map());
      var t = Ce(e.children);
      e.revealOrder && e.revealOrder[0] === "b" && t.reverse();
      for (var r = t.length; r--; ) this.l.set(t[r], (this.i = [1, 0, this.i]));
      return e.children;
    }),
    (lt.prototype.componentDidUpdate = lt.prototype.componentDidMount =
      function () {
        var e = this;
        this.l.forEach(function (t, r) {
          Xo(e, r, t);
        });
      });
  var ia =
      (typeof Symbol < "u" && Symbol.for && Symbol.for("react.element")) ||
      60103,
    ec =
      /^(?:accent|alignment|arabic|baseline|cap|clip(?!PathU)|color|dominant|fill|flood|font|glyph(?!R)|horiz|image(!S)|letter|lighting|marker(?!H|W|U)|overline|paint|pointer|shape|stop|strikethrough|stroke|text(?!L)|transform|underline|unicode|units|v|vector|vert|word|writing|x(?!C))[A-Z]/,
    tc = /^on(Ani|Tra|Tou|BeforeInp|Compo)/,
    rc = /[A-Z0-9]/g,
    nc = typeof document < "u",
    oc = function (e) {
      return (
        typeof Symbol < "u" && typeof Symbol() == "symbol"
          ? /fil|che|rad/
          : /fil|che|ra/
      ).test(e);
    };
  function la(e, t, r) {
    return (
      t.__k == null && (t.textContent = ""),
      Oe(e, t),
      typeof r == "function" && r(),
      e ? e.__c : null
    );
  }
  function ca(e, t, r) {
    return Jr(e, t), typeof r == "function" && r(), e ? e.__c : null;
  }
  (he.prototype.isReactComponent = {}),
    [
      "componentWillMount",
      "componentWillReceiveProps",
      "componentWillUpdate",
    ].forEach(function (e) {
      Object.defineProperty(he.prototype, e, {
        configurable: !0,
        get: function () {
          return this["UNSAFE_" + e];
        },
        set: function (t) {
          Object.defineProperty(this, e, {
            configurable: !0,
            writable: !0,
            value: t,
          });
        },
      });
    });
  var Qo = L.event;
  function ac() {}
  function sc() {
    return this.cancelBubble;
  }
  function ic() {
    return this.defaultPrevented;
  }
  L.event = function (e) {
    return (
      Qo && (e = Qo(e)),
      (e.persist = ac),
      (e.isPropagationStopped = sc),
      (e.isDefaultPrevented = ic),
      (e.nativeEvent = e)
    );
  };
  var vn,
    lc = {
      enumerable: !1,
      configurable: !0,
      get: function () {
        return this.class;
      },
    },
    Ko = L.vnode;
  L.vnode = function (e) {
    typeof e.type == "string" &&
      (function (t) {
        var r = t.props,
          n = t.type,
          a = {},
          s = n.indexOf("-") === -1;
        for (var i in r) {
          var l = r[i];
          if (
            !(
              (i === "value" && "defaultValue" in r && l == null) ||
              (nc && i === "children" && n === "noscript") ||
              i === "class" ||
              i === "className"
            )
          ) {
            var u = i.toLowerCase();
            i === "defaultValue" && "value" in r && r.value == null
              ? (i = "value")
              : i === "download" && l === !0
              ? (l = "")
              : u === "translate" && l === "no"
              ? (l = !1)
              : u[0] === "o" && u[1] === "n"
              ? u === "ondoubleclick"
                ? (i = "ondblclick")
                : u !== "onchange" ||
                  (n !== "input" && n !== "textarea") ||
                  oc(r.type)
                ? u === "onfocus"
                  ? (i = "onfocusin")
                  : u === "onblur"
                  ? (i = "onfocusout")
                  : tc.test(i) && (i = u)
                : (u = i = "oninput")
              : s && ec.test(i)
              ? (i = i.replace(rc, "-$&").toLowerCase())
              : l === null && (l = void 0),
              u === "oninput" && a[(i = u)] && (i = "oninputCapture"),
              (a[i] = l);
          }
        }
        n == "select" &&
          a.multiple &&
          Array.isArray(a.value) &&
          (a.value = Ce(r.children).forEach(function (f) {
            f.props.selected = a.value.indexOf(f.props.value) != -1;
          })),
          n == "select" &&
            a.defaultValue != null &&
            (a.value = Ce(r.children).forEach(function (f) {
              f.props.selected = a.multiple
                ? a.defaultValue.indexOf(f.props.value) != -1
                : a.defaultValue == f.props.value;
            })),
          r.class && !r.className
            ? ((a.class = r.class), Object.defineProperty(a, "className", lc))
            : ((r.className && !r.class) || (r.class && r.className)) &&
              (a.class = a.className = r.className),
          (t.props = a);
      })(e),
      (e.$$typeof = ia),
      Ko && Ko(e);
  };
  var Jo = L.__r;
  L.__r = function (e) {
    Jo && Jo(e), (vn = e.__c);
  };
  var Zo = L.diffed;
  L.diffed = function (e) {
    Zo && Zo(e);
    var t = e.props,
      r = e.__e;
    r != null &&
      e.type === "textarea" &&
      "value" in t &&
      t.value !== r.value &&
      (r.value = t.value == null ? "" : t.value),
      (vn = null);
  };
  var ua = {
      ReactCurrentDispatcher: {
        current: {
          readContext: function (e) {
            return vn.__n[e.__c].props.value;
          },
          useCallback: ce,
          useContext: Fe,
          useDebugValue: rr,
          useDeferredValue: pn,
          useEffect: q,
          useId: nr,
          useImperativeHandle: tr,
          useInsertionEffect: gn,
          useLayoutEffect: Pe,
          useMemo: Z,
          useReducer: He,
          useRef: $,
          useState: P,
          useSyncExternalStore: dn,
          useTransition: hn,
        },
      },
    },
    cc = "18.3.1";
  function fa(e) {
    return g.bind(null, e);
  }
  function Re(e) {
    return !!e && e.$$typeof === ia;
  }
  function da(e) {
    return Re(e) && e.type === h;
  }
  function ma(e) {
    return (
      !!e &&
      !!e.displayName &&
      (typeof e.displayName == "string" || e.displayName instanceof String) &&
      e.displayName.startsWith("Memo(")
    );
  }
  function Ye(e) {
    return Re(e) ? bo.apply(null, arguments) : e;
  }
  function pa(e) {
    return !!e.__k && (Oe(null, e), !0);
  }
  function ha(e) {
    return (e && (e.base || (e.nodeType === 1 && e))) || null;
  }
  var ga = function (e, t) {
      return e(t);
    },
    xn = function (e, t) {
      return e(t);
    },
    va = h,
    xa = Re,
    be = {
      useState: P,
      useId: nr,
      useReducer: He,
      useEffect: q,
      useLayoutEffect: Pe,
      useInsertionEffect: gn,
      useTransition: hn,
      useDeferredValue: pn,
      useSyncExternalStore: dn,
      startTransition: mn,
      useRef: $,
      useImperativeHandle: tr,
      useMemo: Z,
      useCallback: ce,
      useContext: Fe,
      useDebugValue: rr,
      version: "18.3.1",
      Children: Te,
      render: la,
      hydrate: ca,
      unmountComponentAtNode: pa,
      createPortal: sa,
      createElement: g,
      createContext: Ee,
      createFactory: fa,
      cloneElement: Ye,
      createRef: Zt,
      Fragment: h,
      isValidElement: Re,
      isElement: xa,
      isFragment: da,
      isMemo: ma,
      findDOMNode: ha,
      Component: he,
      PureComponent: ir,
      memo: ta,
      forwardRef: H,
      flushSync: xn,
      unstable_batchedUpdates: ga,
      StrictMode: va,
      Suspense: Ct,
      SuspenseList: lt,
      lazy: aa,
      __SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED: ua,
    };
  function ba(e, t) {
    if (typeof e == "function") return e(t);
    e != null && (e.current = t);
  }
  function bn(...e) {
    return (t) => {
      let r = !1,
        n = e.map((a) => {
          let s = ba(a, t);
          return !r && typeof s == "function" && (r = !0), s;
        });
      if (r)
        return () => {
          for (let a = 0; a < n.length; a++) {
            let s = n[a];
            typeof s == "function" ? s() : ba(e[a], null);
          }
        };
    };
  }
  function Xe(...e) {
    return ce(bn(...e), e);
  }
  var uc = 0,
    zd = Array.isArray;
  function o(e, t, r, n, a, s) {
    t || (t = {});
    var i,
      l,
      u = t;
    if ("ref" in u)
      for (l in ((u = {}), t)) l == "ref" ? (i = t[l]) : (u[l] = t[l]);
    var f = {
      type: e,
      props: u,
      key: r,
      ref: i,
      __k: null,
      __: null,
      __b: 0,
      __e: null,
      __c: null,
      constructor: void 0,
      __v: --uc,
      __i: -1,
      __u: 0,
      __source: a,
      __self: s,
    };
    if (typeof e == "function" && (i = e.defaultProps))
      for (l in i) u[l] === void 0 && (u[l] = i[l]);
    return L.vnode && L.vnode(f), f;
  }
  function ct(e) {
    let t = fc(e),
      r = H((n, a) => {
        let { children: s, ...i } = n,
          l = Te.toArray(s),
          u = l.find(mc);
        if (u) {
          let f = u.props.children,
            m = l.map((c) =>
              c === u
                ? Te.count(f) > 1
                  ? Te.only(null)
                  : Re(f)
                  ? f.props.children
                  : null
                : c
            );
          return o(t, {
            ...i,
            ref: a,
            children: Re(f) ? Ye(f, void 0, m) : null,
          });
        }
        return o(t, { ...i, ref: a, children: s });
      });
    return (r.displayName = `${e}.Slot`), r;
  }
  var ya = ct("Slot");
  function fc(e) {
    let t = H((r, n) => {
      let { children: a, ...s } = r;
      if (Re(a)) {
        let i = hc(a),
          l = pc(s, a.props);
        return a.type !== h && (l.ref = n ? bn(n, i) : i), Ye(a, l);
      }
      return Te.count(a) > 1 ? Te.only(null) : null;
    });
    return (t.displayName = `${e}.SlotClone`), t;
  }
  var dc = Symbol("radix.slottable");
  function mc(e) {
    return (
      Re(e) &&
      typeof e.type == "function" &&
      "__radixId" in e.type &&
      e.type.__radixId === dc
    );
  }
  function pc(e, t) {
    let r = { ...t };
    for (let n in t) {
      let a = e[n],
        s = t[n];
      /^on[A-Z]/.test(n)
        ? a && s
          ? (r[n] = (...l) => {
              let u = s(...l);
              return a(...l), u;
            })
          : a && (r[n] = a)
        : n === "style"
        ? (r[n] = { ...a, ...s })
        : n === "className" && (r[n] = [a, s].filter(Boolean).join(" "));
    }
    return { ...e, ...r };
  }
  function hc(e) {
    let t = Object.getOwnPropertyDescriptor(e.props, "ref")?.get,
      r = t && "isReactWarning" in t && t.isReactWarning;
    return r
      ? e.ref
      : ((t = Object.getOwnPropertyDescriptor(e, "ref")?.get),
        (r = t && "isReactWarning" in t && t.isReactWarning),
        r ? e.props.ref : e.props.ref || e.ref);
  }
  var _a = (e) => (typeof e == "boolean" ? `${e}` : e === 0 ? "0" : e),
    wa = or,
    lr = (e, t) => (r) => {
      var n;
      if (t?.variants == null) return wa(e, r?.class, r?.className);
      let { variants: a, defaultVariants: s } = t,
        i = Object.keys(a).map((f) => {
          let m = r?.[f],
            c = s?.[f];
          if (m === null) return null;
          let d = _a(m) || _a(c);
          return a[f][d];
        }),
        l =
          r &&
          Object.entries(r).reduce((f, m) => {
            let [c, d] = m;
            return d === void 0 || (f[c] = d), f;
          }, {}),
        u =
          t == null || (n = t.compoundVariants) === null || n === void 0
            ? void 0
            : n.reduce((f, m) => {
                let { class: c, className: d, ...p } = m;
                return Object.entries(p).every((x) => {
                  let [b, v] = x;
                  return Array.isArray(v)
                    ? v.includes({ ...s, ...l }[b])
                    : { ...s, ...l }[b] === v;
                })
                  ? [...f, c, d]
                  : f;
              }, []);
      return wa(e, i, u, r?.class, r?.className);
    };
  var gc = lr(
      "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium ring-offset-white transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-950 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 dark:ring-offset-slate-950 dark:focus-visible:ring-slate-300",
      {
        variants: {
          variant: {
            default:
              "bg-slate-900 text-slate-50 hover:bg-slate-900/90 dark:bg-slate-50 dark:text-slate-900 dark:hover:bg-slate-50/90",
            destructive:
              "bg-red-500 text-slate-50 hover:bg-red-500/90 dark:bg-red-900 dark:text-slate-50 dark:hover:bg-red-900/90",
            outline:
              "border border-slate-200 bg-white hover:bg-slate-100 hover:text-slate-900 dark:border-slate-800 dark:bg-slate-950 dark:hover:bg-slate-800 dark:hover:text-slate-50",
            secondary:
              "bg-slate-100 text-slate-900 hover:bg-slate-100/80 dark:bg-slate-800 dark:text-slate-50 dark:hover:bg-slate-800/80",
            ghost:
              "hover:bg-slate-100 hover:text-slate-900 dark:hover:bg-slate-800 dark:hover:text-slate-50",
            link: "text-slate-900 underline-offset-4 hover:underline dark:text-slate-50",
          },
          size: {
            default: "h-10 px-4 py-2",
            sm: "h-9 rounded-md px-3",
            lg: "h-11 rounded-md px-8",
            icon: "h-10 w-10",
          },
        },
        defaultVariants: { variant: "default", size: "default" },
      }
    ),
    K = H(({ className: e, variant: t, size: r, asChild: n = !1, ...a }, s) =>
      o(n ? ya : "button", {
        className: D(gc({ variant: t, size: r, className: e })),
        ref: s,
        ...a,
      })
    );
  K.displayName = "Button";
  var vc = [
    { id: "dashboard", label: "Dashboard" },
    { id: "requests", label: "Requests" },
    { id: "analytics", label: "Analytics" },
    { id: "environment", label: "Environment" },
    { id: "trace", label: "Trace" },
  ];
  function Na({ activeTab: e, onTabChange: t, stats: r, onClearAll: n }) {
    return o("aside", {
      className: "w-64 bg-white border-r h-screen flex flex-col shadow-sm",
      children: [
        o("div", {
          className: "px-6 py-5 border-b",
          children: [
            o("h1", { className: "text-xl font-bold", children: "GoVisual" }),
            o("p", {
              className: "text-xs text-muted-foreground mt-1",
              children: "HTTP Request Visualizer",
            }),
          ],
        }),
        o("nav", {
          className: "flex-1 px-4 py-6",
          children: o("div", {
            className: "space-y-1",
            children: vc.map((a) =>
              o(
                "button",
                {
                  onClick: () => t(a.id),
                  className: D(
                    "w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200",
                    e === a.id
                      ? "bg-slate-900 text-white shadow-sm"
                      : "text-slate-600 hover:bg-slate-100 hover:text-slate-900"
                  ),
                  children: [
                    o("span", { children: a.label }),
                    e === a.id &&
                      o("span", {
                        className: "ml-auto w-1.5 h-1.5 bg-white rounded-full",
                      }),
                  ],
                },
                a.id
              )
            ),
          }),
        }),
        o("div", {
          className: "px-4 py-4 border-t bg-slate-50/50",
          children: [
            o("h3", {
              className:
                "text-xs font-semibold text-slate-500 uppercase mb-3 tracking-wider",
              children: "Quick Stats",
            }),
            o("div", {
              className: "space-y-2",
              children: [
                o("div", {
                  className:
                    "flex justify-between items-baseline px-3 py-2 rounded-lg hover:bg-white transition-colors",
                  children: [
                    o("span", {
                      className: "text-xs text-slate-500",
                      children: "Total",
                    }),
                    o("span", {
                      className: "text-sm font-bold tabular-nums",
                      children: r.total,
                    }),
                  ],
                }),
                o("div", {
                  className:
                    "flex justify-between items-baseline px-3 py-2 rounded-lg hover:bg-white transition-colors",
                  children: [
                    o("span", {
                      className: "text-xs text-slate-500",
                      children: "Success",
                    }),
                    o("span", {
                      className: "text-sm font-bold tabular-nums",
                      children: [r.successRate, "%"],
                    }),
                  ],
                }),
                o("div", {
                  className:
                    "flex justify-between items-baseline px-3 py-2 rounded-lg hover:bg-white transition-colors",
                  children: [
                    o("span", {
                      className: "text-xs text-slate-500",
                      children: "Avg",
                    }),
                    o("span", {
                      className: "text-sm font-bold tabular-nums",
                      children: [r.avgDuration, "ms"],
                    }),
                  ],
                }),
              ],
            }),
          ],
        }),
        o("div", {
          className: "p-4 border-t space-y-3",
          children: [
            o(K, {
              variant: "destructive",
              className: "w-full",
              onClick: n,
              children: "Clear All Requests",
            }),
            o("div", {
              className: "text-center",
              children: [
                o("p", {
                  className: "text-[10px] text-slate-500 font-medium",
                  children: "VERSION 0.2.0",
                }),
                o("p", {
                  className: "text-[10px] text-slate-500 mt-1",
                  children: [
                    "Created by",
                    " ",
                    o("a", {
                      href: "https://github.com/doganarif",
                      target: "_blank",
                      rel: "noopener noreferrer",
                      className: "text-slate-900 hover:underline font-medium",
                      children: "@doganarif",
                    }),
                  ],
                }),
              ],
            }),
          ],
        }),
      ],
    });
  }
  var I = H(({ className: e, ...t }, r) =>
    o("div", {
      ref: r,
      className: D(
        "rounded-lg border border-slate-200 bg-white text-slate-950 shadow-sm",
        e
      ),
      ...t,
    })
  );
  I.displayName = "Card";
  var j = H(({ className: e, ...t }, r) =>
    o("div", { ref: r, className: D("flex flex-col space-y-1.5 p-6", e), ...t })
  );
  j.displayName = "CardHeader";
  var W = H(({ className: e, ...t }, r) =>
    o("div", {
      ref: r,
      className: D("text-2xl font-semibold leading-none tracking-tight", e),
      ...t,
    })
  );
  W.displayName = "CardTitle";
  var Tt = H(({ className: e, ...t }, r) =>
    o("div", { ref: r, className: D("text-sm text-slate-500", e), ...t })
  );
  Tt.displayName = "CardDescription";
  var A = H(({ className: e, ...t }, r) =>
    o("div", { ref: r, className: D("p-6 pt-0", e), ...t })
  );
  A.displayName = "CardContent";
  var xc = H(({ className: e, ...t }, r) =>
    o("div", { ref: r, className: D("flex items-center p-6 pt-0", e), ...t })
  );
  xc.displayName = "CardFooter";
  function Ca({ requests: e }) {
    let r = (() => {
        let a = e.length,
          s = e.filter((m) => m.StatusCode >= 200 && m.StatusCode < 300).length,
          i = e.filter((m) => m.StatusCode >= 300 && m.StatusCode < 400).length,
          l = e.filter((m) => m.StatusCode >= 400 && m.StatusCode < 500).length,
          u = e.filter((m) => m.StatusCode >= 500).length,
          f = a > 0 ? Math.round(e.reduce((m, c) => m + c.Duration, 0) / a) : 0;
        return {
          total: a,
          success: s,
          redirect: i,
          clientError: l,
          serverError: u,
          avgDuration: f,
        };
      })(),
      n = [
        { label: "Total Requests", value: r.total },
        { label: "Success (2xx)", value: r.success },
        { label: "Redirect (3xx)", value: r.redirect },
        { label: "Client Error (4xx)", value: r.clientError },
        { label: "Server Error (5xx)", value: r.serverError },
        { label: "Avg Response Time", value: `${r.avgDuration}ms` },
      ];
    return o("div", {
      className: "grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-6",
      children: n.map((a, s) =>
        o(
          I,
          {
            className:
              "hover:shadow-lg transition-all duration-300 hover:-translate-y-1 cursor-default",
            style: { animationDelay: `${s * 50}ms` },
            children: o(A, {
              className: "p-6",
              children: [
                o("div", {
                  className: "text-3xl font-bold mb-2 tabular-nums",
                  children: a.value,
                }),
                o("p", {
                  className:
                    "text-xs text-muted-foreground font-medium uppercase tracking-wider",
                  children: a.label,
                }),
              ],
            }),
          },
          s
        )
      ),
    });
  }
  var Rt = H(({ className: e, ...t }, r) =>
    o("div", {
      className: "relative w-full overflow-auto",
      children: o("table", {
        ref: r,
        className: D("w-full caption-bottom text-sm", e),
        ...t,
      }),
    })
  );
  Rt.displayName = "Table";
  var kt = H(({ className: e, ...t }, r) =>
    o("thead", { ref: r, className: D("[&_tr]:border-b", e), ...t })
  );
  kt.displayName = "TableHeader";
  var St = H(({ className: e, ...t }, r) =>
    o("tbody", { ref: r, className: D("[&_tr:last-child]:border-0", e), ...t })
  );
  St.displayName = "TableBody";
  var bc = H(({ className: e, ...t }, r) =>
    o("tfoot", {
      ref: r,
      className: D(
        "border-t bg-slate-100/50 font-medium [&>tr]:last:border-b-0 dark:bg-slate-800/50",
        e
      ),
      ...t,
    })
  );
  bc.displayName = "TableFooter";
  var Qe = H(({ className: e, ...t }, r) =>
    o("tr", {
      ref: r,
      className: D(
        "border-b transition-colors hover:bg-slate-100/50 data-[state=selected]:bg-slate-100 dark:hover:bg-slate-800/50 dark:data-[state=selected]:bg-slate-800",
        e
      ),
      ...t,
    })
  );
  Qe.displayName = "TableRow";
  var ge = H(({ className: e, ...t }, r) =>
    o("th", {
      ref: r,
      className: D(
        "h-12 px-4 text-left align-middle font-medium text-slate-500 [&:has([role=checkbox])]:pr-0 dark:text-slate-400",
        e
      ),
      ...t,
    })
  );
  ge.displayName = "TableHead";
  var ve = H(({ className: e, ...t }, r) =>
    o("td", {
      ref: r,
      className: D("p-4 align-middle [&:has([role=checkbox])]:pr-0", e),
      ...t,
    })
  );
  ve.displayName = "TableCell";
  var yc = H(({ className: e, ...t }, r) =>
    o("caption", {
      ref: r,
      className: D("mt-4 text-sm text-slate-500 dark:text-slate-400", e),
      ...t,
    })
  );
  yc.displayName = "TableCaption";
  var _c = lr(
    "inline-flex items-center rounded-full border border-slate-200 px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-slate-950 focus:ring-offset-2",
    {
      variants: {
        variant: {
          default:
            "border-transparent bg-slate-900 text-slate-50 hover:bg-slate-900/80",
          secondary:
            "border-transparent bg-slate-100 text-slate-900 hover:bg-slate-100/80",
          destructive:
            "border-transparent bg-red-500 text-slate-50 hover:bg-red-500/80",
          outline: "text-slate-950",
        },
      },
      defaultVariants: { variant: "default" },
    }
  );
  function ie({ className: e, variant: t, ...r }) {
    return o("div", { className: D(_c({ variant: t }), e), ...r });
  }
  function cr({
    requests: e,
    selectedRequest: t,
    onRequestSelect: r,
    selectedForComparison: n = [],
    onToggleComparison: a,
    onReplay: s,
  }) {
    let i = (c) =>
        c >= 200 && c < 300
          ? "default"
          : c >= 300 && c < 400
          ? "secondary"
          : c >= 400
          ? "outline"
          : "secondary",
      l = (c) => `method-${c.toLowerCase()}`,
      u = (c) =>
        c < 1 ? "<1ms" : c < 1e3 ? `${c}ms` : `${(c / 1e3).toFixed(2)}s`,
      f = (c) => new Date(c).toLocaleTimeString(),
      m = (c) => (c > 500 ? "font-semibold" : c > 200 ? "font-medium" : "");
    return !e || e.length === 0
      ? o("div", {
          className:
            "flex flex-col items-center justify-center h-64 text-muted-foreground",
          children: [
            o("svg", {
              className: "w-16 h-16 mb-4 opacity-30",
              fill: "none",
              stroke: "currentColor",
              viewBox: "0 0 24 24",
              children: o("path", {
                strokeLinecap: "round",
                strokeLinejoin: "round",
                strokeWidth: 1.5,
                d: "M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z",
              }),
            }),
            o("p", {
              className: "text-sm font-medium",
              children: "No requests logged yet",
            }),
            o("p", {
              className: "text-xs text-muted-foreground/70 mt-1",
              children: "Waiting for incoming HTTP requests...",
            }),
          ],
        })
      : o(Rt, {
          children: [
            o(kt, {
              children: o(Qe, {
                children: [
                  a &&
                    o(ge, {
                      className: "w-12",
                      children: o("span", {
                        className: "sr-only",
                        children: "Select",
                      }),
                    }),
                  o(ge, { className: "w-24", children: "Time" }),
                  o(ge, { className: "w-20", children: "Method" }),
                  o(ge, { children: "Path" }),
                  o(ge, { className: "w-20", children: "Status" }),
                  o(ge, { className: "w-24 text-right", children: "Duration" }),
                  s &&
                    o(ge, {
                      className: "w-20",
                      children: o("span", {
                        className: "sr-only",
                        children: "Actions",
                      }),
                    }),
                ],
              }),
            }),
            o(St, {
              children: e.map((c) =>
                o(
                  Qe,
                  {
                    onClick: (d) => {
                      let p = d.target;
                      p.tagName !== "INPUT" &&
                        p.tagName !== "BUTTON" &&
                        !p.closest("button") &&
                        r(c);
                    },
                    className: D(
                      "cursor-pointer transition-all duration-150 hover:bg-muted/50",
                      t?.ID === c.ID && "bg-accent/50 shadow-sm"
                    ),
                    style: { animationDelay: `${e.indexOf(c) * 20}ms` },
                    children: [
                      a &&
                        o(ve, {
                          children: o("input", {
                            type: "checkbox",
                            checked: n.includes(c.ID),
                            onChange: (d) => {
                              d.stopPropagation(), a(c.ID);
                            },
                            className: "h-4 w-4 rounded border-gray-300",
                            onClick: (d) => d.stopPropagation(),
                          }),
                        }),
                      o(ve, {
                        className: "text-xs text-muted-foreground font-medium",
                        children: f(c.Timestamp),
                      }),
                      o(ve, {
                        children: o("span", {
                          className: D(
                            "px-2 py-1 rounded text-xs",
                            l(c.Method)
                          ),
                          children: c.Method,
                        }),
                      }),
                      o(ve, {
                        className: "font-mono text-sm",
                        children: [
                          o("span", {
                            className: "font-medium",
                            children: c.Path,
                          }),
                          c.Query &&
                            o("span", {
                              className: "text-muted-foreground text-xs",
                              children: ["?", c.Query],
                            }),
                        ],
                      }),
                      o(ve, {
                        children: o(ie, {
                          variant: i(c.StatusCode),
                          className: "font-mono",
                          children: c.StatusCode,
                        }),
                      }),
                      o(ve, {
                        className: D(
                          "text-right tabular-nums font-mono text-sm",
                          m(c.Duration)
                        ),
                        children: u(c.Duration),
                      }),
                      s &&
                        o(ve, {
                          children: o(K, {
                            size: "sm",
                            variant: "ghost",
                            onClick: (d) => {
                              d.stopPropagation(), s(c);
                            },
                            className: "h-7 px-2 text-xs",
                            children: "Replay",
                          }),
                        }),
                    ],
                  },
                  c.ID
                )
              ),
            }),
          ],
        });
  }
  function Ta({ open: e, onOpenChange: t, children: r, className: n }) {
    return (
      q(
        () => (
          e
            ? (document.body.style.overflow = "hidden")
            : (document.body.style.overflow = ""),
          () => {
            document.body.style.overflow = "";
          }
        ),
        [e]
      ),
      e
        ? o("div", {
            className: D("drawer-container", n),
            children: [
              o("div", {
                className: D(
                  "fixed inset-0 bg-black/50 backdrop-blur-sm z-40",
                  "animate-in fade-in duration-200"
                ),
                onClick: () => t(!1),
              }),
              o("div", {
                className: "fixed inset-0 z-50 pointer-events-none",
                children: r,
              }),
            ],
          })
        : null
    );
  }
  function Ra({
    children: e,
    className: t,
    onClose: r,
    isFullscreen: n,
    onToggleFullscreen: a,
  }) {
    let s = $(null),
      i = $(null),
      l = $(0),
      u = $(0),
      f = $(!1);
    return (
      q(() => {
        let m = (x) => {
            if (!f.current || n) return;
            let b = x.clientY - l.current;
            (u.current = Math.max(0, b)),
              s.current &&
                (s.current.style.transform = `translateY(${u.current}px)`);
          },
          c = () => {
            f.current &&
              ((f.current = !1),
              u.current > 200
                ? r()
                : s.current &&
                  ((s.current.style.transform = ""),
                  (s.current.style.transition = "transform 0.3s ease"),
                  setTimeout(() => {
                    s.current && (s.current.style.transition = "");
                  }, 300)));
          },
          d = (x) => {
            n || ((f.current = !0), (l.current = x.clientY), (u.current = 0));
          },
          p = i.current;
        return (
          p && p.addEventListener("mousedown", d),
          document.addEventListener("mousemove", m),
          document.addEventListener("mouseup", c),
          () => {
            p && p.removeEventListener("mousedown", d),
              document.removeEventListener("mousemove", m),
              document.removeEventListener("mouseup", c);
          }
        );
      }, [n, r]),
      o("div", {
        ref: s,
        className: D(
          "fixed bg-background rounded-t-2xl shadow-2xl pointer-events-auto",
          "animate-in slide-in-from-bottom duration-300",
          n ? "inset-0 rounded-none" : "inset-x-0 bottom-0 max-h-[85vh]",
          t
        ),
        children: [
          !n &&
            o("div", {
              ref: i,
              className:
                "absolute top-0 left-0 right-0 h-8 cursor-ns-resize flex items-center justify-center",
              children: o("div", {
                className: "w-12 h-1 bg-muted-foreground/30 rounded-full",
              }),
            }),
          o("div", {
            className:
              "sticky top-0 bg-background/95 backdrop-blur-sm border-b z-10 px-6 py-4",
            children: o("div", {
              className: "flex items-center justify-between",
              children: [
                o("h3", {
                  className: "text-lg font-semibold",
                  children: "Request Details",
                }),
                o("div", {
                  className: "flex items-center gap-2",
                  children: [
                    o("button", {
                      onClick: a,
                      className:
                        "p-2 rounded-lg hover:bg-accent transition-colors",
                      "aria-label": n ? "Exit fullscreen" : "Enter fullscreen",
                      children: n
                        ? o("svg", {
                            className: "w-4 h-4",
                            fill: "none",
                            stroke: "currentColor",
                            viewBox: "0 0 24 24",
                            children: o("path", {
                              strokeLinecap: "round",
                              strokeLinejoin: "round",
                              strokeWidth: 2,
                              d: "M6 18L18 6M6 6l12 12",
                            }),
                          })
                        : o("svg", {
                            className: "w-4 h-4",
                            fill: "none",
                            stroke: "currentColor",
                            viewBox: "0 0 24 24",
                            children: o("path", {
                              strokeLinecap: "round",
                              strokeLinejoin: "round",
                              strokeWidth: 2,
                              d: "M4 8V4m0 0h4M4 4l5 5m11-5h-4m4 0v4m0 0l-5-5M4 16v4m0 0h4M4 20l5-5m11 5h-4m4 0v-4m0 0l-5 5",
                            }),
                          }),
                    }),
                    o("button", {
                      onClick: r,
                      className:
                        "p-2 rounded-lg hover:bg-accent transition-colors",
                      "aria-label": "Close drawer",
                      children: o("svg", {
                        className: "w-4 h-4",
                        fill: "none",
                        stroke: "currentColor",
                        viewBox: "0 0 24 24",
                        children: o("path", {
                          strokeLinecap: "round",
                          strokeLinejoin: "round",
                          strokeWidth: 2,
                          d: "M6 18L18 6M6 6l12 12",
                        }),
                      }),
                    }),
                  ],
                }),
              ],
            }),
          }),
          o("div", {
            className: D(
              "overflow-y-auto",
              n ? "h-[calc(100vh-4rem)]" : "max-h-[calc(85vh-4rem)]"
            ),
            children: o("div", { className: "p-6", children: e }),
          }),
        ],
      })
    );
  }
  var zm = !!(
    typeof window < "u" &&
    window.document &&
    window.document.createElement
  );
  function ye(e, t, { checkForDefaultPrevented: r = !0 } = {}) {
    return function (a) {
      if ((e?.(a), r === !1 || !a.defaultPrevented)) return t?.(a);
    };
  }
  function ut(e, t = []) {
    let r = [];
    function n(s, i) {
      let l = Ee(i),
        u = r.length;
      r = [...r, i];
      let f = (c) => {
        let { scope: d, children: p, ...x } = c,
          b = d?.[e]?.[u] || l,
          v = Z(() => x, Object.values(x));
        return o(b.Provider, { value: v, children: p });
      };
      f.displayName = s + "Provider";
      function m(c, d) {
        let p = d?.[e]?.[u] || l,
          x = Fe(p);
        if (x) return x;
        if (i !== void 0) return i;
        throw new Error(`\`${c}\` must be used within \`${s}\``);
      }
      return [f, m];
    }
    let a = () => {
      let s = r.map((i) => Ee(i));
      return function (l) {
        let u = l?.[e] || s;
        return Z(() => ({ [`__scope${e}`]: { ...l, [e]: u } }), [l, u]);
      };
    };
    return (a.scopeName = e), [n, wc(a, ...t)];
  }
  function wc(...e) {
    let t = e[0];
    if (e.length === 1) return t;
    let r = () => {
      let n = e.map((a) => ({ useScope: a(), scopeName: a.scopeName }));
      return function (s) {
        let i = n.reduce((l, { useScope: u, scopeName: f }) => {
          let c = u(s)[`__scope${f}`];
          return { ...l, ...c };
        }, {});
        return Z(() => ({ [`__scope${t.scopeName}`]: i }), [i]);
      };
    };
    return (r.scopeName = t.scopeName), r;
  }
  function ka(e) {
    let t = e + "CollectionProvider",
      [r, n] = ut(t),
      [a, s] = r(t, { collectionRef: { current: null }, itemMap: new Map() }),
      i = (b) => {
        let { scope: v, children: y } = b,
          N = be.useRef(null),
          R = be.useRef(new Map()).current;
        return o(a, { scope: v, itemMap: R, collectionRef: N, children: y });
      };
    i.displayName = t;
    let l = e + "CollectionSlot",
      u = ct(l),
      f = be.forwardRef((b, v) => {
        let { scope: y, children: N } = b,
          R = s(l, y),
          M = Xe(v, R.collectionRef);
        return o(u, { ref: M, children: N });
      });
    f.displayName = l;
    let m = e + "CollectionItemSlot",
      c = "data-radix-collection-item",
      d = ct(m),
      p = be.forwardRef((b, v) => {
        let { scope: y, children: N, ...R } = b,
          M = be.useRef(null),
          w = Xe(v, M),
          _ = s(m, y);
        return (
          be.useEffect(
            () => (
              _.itemMap.set(M, { ref: M, ...R }), () => void _.itemMap.delete(M)
            )
          ),
          o(d, { [c]: "", ref: w, children: N })
        );
      });
    p.displayName = m;
    function x(b) {
      let v = s(e + "CollectionConsumer", b);
      return be.useCallback(() => {
        let N = v.collectionRef.current;
        if (!N) return [];
        let R = Array.from(N.querySelectorAll(`[${c}]`));
        return Array.from(v.itemMap.values()).sort(
          (_, T) => R.indexOf(_.ref.current) - R.indexOf(T.ref.current)
        );
      }, [v.collectionRef, v.itemMap]);
    }
    return [{ Provider: i, Slot: f, ItemSlot: p }, x, n];
  }
  var Ke = globalThis?.document ? Pe : () => {};
  var Nc = ee[" useId ".trim().toString()] || (() => {}),
    Cc = 0;
  function ur(e) {
    let [t, r] = P(Nc());
    return (
      Ke(() => {
        e || r((n) => n ?? String(Cc++));
      }, [e]),
      e || (t ? `radix-${t}` : "")
    );
  }
  var Tc = [
      "a",
      "button",
      "div",
      "form",
      "h2",
      "h3",
      "img",
      "input",
      "label",
      "li",
      "nav",
      "ol",
      "p",
      "select",
      "span",
      "svg",
      "ul",
    ],
    ze = Tc.reduce((e, t) => {
      let r = ct(`Primitive.${t}`),
        n = H((a, s) => {
          let { asChild: i, ...l } = a,
            u = i ? r : t;
          return (
            typeof window < "u" && (window[Symbol.for("radix-ui")] = !0),
            o(u, { ...l, ref: s })
          );
        });
      return (n.displayName = `Primitive.${t}`), { ...e, [t]: n };
    }, {});
  function Sa(e) {
    let t = $(e);
    return (
      q(() => {
        t.current = e;
      }),
      Z(
        () =>
          (...r) =>
            t.current?.(...r),
        []
      )
    );
  }
  var Rc = ee[" useInsertionEffect ".trim().toString()] || Ke;
  function fr({ prop: e, defaultProp: t, onChange: r = () => {}, caller: n }) {
    let [a, s, i] = kc({ defaultProp: t, onChange: r }),
      l = e !== void 0,
      u = l ? e : a;
    {
      let m = $(e !== void 0);
      q(() => {
        let c = m.current;
        c !== l &&
          console.warn(
            `${n} is changing from ${c ? "controlled" : "uncontrolled"} to ${
              l ? "controlled" : "uncontrolled"
            }. Components should not switch from controlled to uncontrolled (or vice versa). Decide between using a controlled or uncontrolled value for the lifetime of the component.`
          ),
          (m.current = l);
      }, [l, n]);
    }
    let f = ce(
      (m) => {
        if (l) {
          let c = Sc(m) ? m(e) : m;
          c !== e && i.current?.(c);
        } else s(m);
      },
      [l, e, s, i]
    );
    return [u, f];
  }
  function kc({ defaultProp: e, onChange: t }) {
    let [r, n] = P(e),
      a = $(r),
      s = $(t);
    return (
      Rc(() => {
        s.current = t;
      }, [t]),
      q(() => {
        a.current !== r && (s.current?.(r), (a.current = r));
      }, [r, a]),
      [r, n, s]
    );
  }
  function Sc(e) {
    return typeof e == "function";
  }
  var vp = Symbol("RADIX:SYNC_STATE");
  var Ec = Ee(void 0);
  function dr(e) {
    let t = Fe(Ec);
    return e || t || "ltr";
  }
  var yn = "rovingFocusGroup.onEntryFocus",
    Mc = { bubbles: !1, cancelable: !0 },
    Et = "RovingFocusGroup",
    [_n, Ea, Pc] = ka(Et),
    [Ic, wn] = ut(Et, [Pc]),
    [Ac, Dc] = Ic(Et),
    Ma = H((e, t) =>
      o(_n.Provider, {
        scope: e.__scopeRovingFocusGroup,
        children: o(_n.Slot, {
          scope: e.__scopeRovingFocusGroup,
          children: o(Lc, { ...e, ref: t }),
        }),
      })
    );
  Ma.displayName = Et;
  var Lc = H((e, t) => {
      let {
          __scopeRovingFocusGroup: r,
          orientation: n,
          loop: a = !1,
          dir: s,
          currentTabStopId: i,
          defaultCurrentTabStopId: l,
          onCurrentTabStopIdChange: u,
          onEntryFocus: f,
          preventScrollOnEntryFocus: m = !1,
          ...c
        } = e,
        d = $(null),
        p = Xe(t, d),
        x = dr(s),
        [b, v] = fr({
          prop: i,
          defaultProp: l ?? null,
          onChange: u,
          caller: Et,
        }),
        [y, N] = P(!1),
        R = Sa(f),
        M = Ea(r),
        w = $(!1),
        [_, T] = P(0);
      return (
        q(() => {
          let C = d.current;
          if (C)
            return (
              C.addEventListener(yn, R), () => C.removeEventListener(yn, R)
            );
        }, [R]),
        o(Ac, {
          scope: r,
          orientation: n,
          dir: x,
          loop: a,
          currentTabStopId: b,
          onItemFocus: ce((C) => v(C), [v]),
          onItemShiftTab: ce(() => N(!0), []),
          onFocusableItemAdd: ce(() => T((C) => C + 1), []),
          onFocusableItemRemove: ce(() => T((C) => C - 1), []),
          children: o(ze.div, {
            tabIndex: y || _ === 0 ? -1 : 0,
            "data-orientation": n,
            ...c,
            ref: p,
            style: { outline: "none", ...e.style },
            onMouseDown: ye(e.onMouseDown, () => {
              w.current = !0;
            }),
            onFocus: ye(e.onFocus, (C) => {
              let V = !w.current;
              if (C.target === C.currentTarget && V && !y) {
                let U = new CustomEvent(yn, Mc);
                if ((C.currentTarget.dispatchEvent(U), !U.defaultPrevented)) {
                  let B = M().filter((O) => O.focusable),
                    oe = B.find((O) => O.active),
                    rt = B.find((O) => O.id === b),
                    z = [oe, rt, ...B]
                      .filter(Boolean)
                      .map((O) => O.ref.current);
                  Aa(z, m);
                }
              }
              w.current = !1;
            }),
            onBlur: ye(e.onBlur, () => N(!1)),
          }),
        })
      );
    }),
    Pa = "RovingFocusGroupItem",
    Ia = H((e, t) => {
      let {
          __scopeRovingFocusGroup: r,
          focusable: n = !0,
          active: a = !1,
          tabStopId: s,
          children: i,
          ...l
        } = e,
        u = ur(),
        f = s || u,
        m = Dc(Pa, r),
        c = m.currentTabStopId === f,
        d = Ea(r),
        {
          onFocusableItemAdd: p,
          onFocusableItemRemove: x,
          currentTabStopId: b,
        } = m;
      return (
        q(() => {
          if (n) return p(), () => x();
        }, [n, p, x]),
        o(_n.ItemSlot, {
          scope: r,
          id: f,
          focusable: n,
          active: a,
          children: o(ze.span, {
            tabIndex: c ? 0 : -1,
            "data-orientation": m.orientation,
            ...l,
            ref: t,
            onMouseDown: ye(e.onMouseDown, (v) => {
              n ? m.onItemFocus(f) : v.preventDefault();
            }),
            onFocus: ye(e.onFocus, () => m.onItemFocus(f)),
            onKeyDown: ye(e.onKeyDown, (v) => {
              if (v.key === "Tab" && v.shiftKey) {
                m.onItemShiftTab();
                return;
              }
              if (v.target !== v.currentTarget) return;
              let y = Fc(v, m.orientation, m.dir);
              if (y !== void 0) {
                if (v.metaKey || v.ctrlKey || v.altKey || v.shiftKey) return;
                v.preventDefault();
                let R = d()
                  .filter((M) => M.focusable)
                  .map((M) => M.ref.current);
                if (y === "last") R.reverse();
                else if (y === "prev" || y === "next") {
                  y === "prev" && R.reverse();
                  let M = R.indexOf(v.currentTarget);
                  R = m.loop ? Bc(R, M + 1) : R.slice(M + 1);
                }
                setTimeout(() => Aa(R));
              }
            }),
            children:
              typeof i == "function"
                ? i({ isCurrentTabStop: c, hasTabStop: b != null })
                : i,
          }),
        })
      );
    });
  Ia.displayName = Pa;
  var Oc = {
    ArrowLeft: "prev",
    ArrowUp: "prev",
    ArrowRight: "next",
    ArrowDown: "next",
    PageUp: "first",
    Home: "first",
    PageDown: "last",
    End: "last",
  };
  function Hc(e, t) {
    return t !== "rtl"
      ? e
      : e === "ArrowLeft"
      ? "ArrowRight"
      : e === "ArrowRight"
      ? "ArrowLeft"
      : e;
  }
  function Fc(e, t, r) {
    let n = Hc(e.key, r);
    if (
      !(t === "vertical" && ["ArrowLeft", "ArrowRight"].includes(n)) &&
      !(t === "horizontal" && ["ArrowUp", "ArrowDown"].includes(n))
    )
      return Oc[n];
  }
  function Aa(e, t = !1) {
    let r = document.activeElement;
    for (let n of e)
      if (
        n === r ||
        (n.focus({ preventScroll: t }), document.activeElement !== r)
      )
        return;
  }
  function Bc(e, t) {
    return e.map((r, n) => e[(t + n) % e.length]);
  }
  var Da = Ma,
    La = Ia;
  function qc(e, t) {
    return He((r, n) => t[r][n] ?? r, e);
  }
  var Nn = (e) => {
    let { present: t, children: r } = e,
      n = $c(t),
      a = typeof r == "function" ? r({ present: n.isPresent }) : Te.only(r),
      s = Xe(n.ref, Vc(a));
    return typeof r == "function" || n.isPresent ? Ye(a, { ref: s }) : null;
  };
  Nn.displayName = "Presence";
  function $c(e) {
    let [t, r] = P(),
      n = $(null),
      a = $(e),
      s = $("none"),
      i = e ? "mounted" : "unmounted",
      [l, u] = qc(i, {
        mounted: { UNMOUNT: "unmounted", ANIMATION_OUT: "unmountSuspended" },
        unmountSuspended: { MOUNT: "mounted", ANIMATION_END: "unmounted" },
        unmounted: { MOUNT: "mounted" },
      });
    return (
      q(() => {
        let f = mr(n.current);
        s.current = l === "mounted" ? f : "none";
      }, [l]),
      Ke(() => {
        let f = n.current,
          m = a.current;
        if (m !== e) {
          let d = s.current,
            p = mr(f);
          e
            ? u("MOUNT")
            : p === "none" || f?.display === "none"
            ? u("UNMOUNT")
            : u(m && d !== p ? "ANIMATION_OUT" : "UNMOUNT"),
            (a.current = e);
        }
      }, [e, u]),
      Ke(() => {
        if (t) {
          let f,
            m = t.ownerDocument.defaultView ?? window,
            c = (p) => {
              let b = mr(n.current).includes(CSS.escape(p.animationName));
              if (p.target === t && b && (u("ANIMATION_END"), !a.current)) {
                let v = t.style.animationFillMode;
                (t.style.animationFillMode = "forwards"),
                  (f = m.setTimeout(() => {
                    t.style.animationFillMode === "forwards" &&
                      (t.style.animationFillMode = v);
                  }));
              }
            },
            d = (p) => {
              p.target === t && (s.current = mr(n.current));
            };
          return (
            t.addEventListener("animationstart", d),
            t.addEventListener("animationcancel", c),
            t.addEventListener("animationend", c),
            () => {
              m.clearTimeout(f),
                t.removeEventListener("animationstart", d),
                t.removeEventListener("animationcancel", c),
                t.removeEventListener("animationend", c);
            }
          );
        } else u("ANIMATION_END");
      }, [t, u]),
      {
        isPresent: ["mounted", "unmountSuspended"].includes(l),
        ref: ce((f) => {
          (n.current = f ? getComputedStyle(f) : null), r(f);
        }, []),
      }
    );
  }
  function mr(e) {
    return e?.animationName || "none";
  }
  function Vc(e) {
    let t = Object.getOwnPropertyDescriptor(e.props, "ref")?.get,
      r = t && "isReactWarning" in t && t.isReactWarning;
    return r
      ? e.ref
      : ((t = Object.getOwnPropertyDescriptor(e, "ref")?.get),
        (r = t && "isReactWarning" in t && t.isReactWarning),
        r ? e.props.ref : e.props.ref || e.ref);
  }
  var pr = "Tabs",
    [Gc, Wp] = ut(pr, [wn]),
    Oa = wn(),
    [Uc, Cn] = Gc(pr),
    Ha = H((e, t) => {
      let {
          __scopeTabs: r,
          value: n,
          onValueChange: a,
          defaultValue: s,
          orientation: i = "horizontal",
          dir: l,
          activationMode: u = "automatic",
          ...f
        } = e,
        m = dr(l),
        [c, d] = fr({ prop: n, onChange: a, defaultProp: s ?? "", caller: pr });
      return o(Uc, {
        scope: r,
        baseId: ur(),
        value: c,
        onValueChange: d,
        orientation: i,
        dir: m,
        activationMode: u,
        children: o(ze.div, { dir: m, "data-orientation": i, ...f, ref: t }),
      });
    });
  Ha.displayName = pr;
  var Fa = "TabsList",
    Ba = H((e, t) => {
      let { __scopeTabs: r, loop: n = !0, ...a } = e,
        s = Cn(Fa, r),
        i = Oa(r);
      return o(Da, {
        asChild: !0,
        ...i,
        orientation: s.orientation,
        dir: s.dir,
        loop: n,
        children: o(ze.div, {
          role: "tablist",
          "aria-orientation": s.orientation,
          ...a,
          ref: t,
        }),
      });
    });
  Ba.displayName = Fa;
  var za = "TabsTrigger",
    qa = H((e, t) => {
      let { __scopeTabs: r, value: n, disabled: a = !1, ...s } = e,
        i = Cn(za, r),
        l = Oa(r),
        u = Ga(i.baseId, n),
        f = Ua(i.baseId, n),
        m = n === i.value;
      return o(La, {
        asChild: !0,
        ...l,
        focusable: !a,
        active: m,
        children: o(ze.button, {
          type: "button",
          role: "tab",
          "aria-selected": m,
          "aria-controls": f,
          "data-state": m ? "active" : "inactive",
          "data-disabled": a ? "" : void 0,
          disabled: a,
          id: u,
          ...s,
          ref: t,
          onMouseDown: ye(e.onMouseDown, (c) => {
            !a && c.button === 0 && c.ctrlKey === !1
              ? i.onValueChange(n)
              : c.preventDefault();
          }),
          onKeyDown: ye(e.onKeyDown, (c) => {
            [" ", "Enter"].includes(c.key) && i.onValueChange(n);
          }),
          onFocus: ye(e.onFocus, () => {
            let c = i.activationMode !== "manual";
            !m && !a && c && i.onValueChange(n);
          }),
        }),
      });
    });
  qa.displayName = za;
  var $a = "TabsContent",
    Va = H((e, t) => {
      let { __scopeTabs: r, value: n, forceMount: a, children: s, ...i } = e,
        l = Cn($a, r),
        u = Ga(l.baseId, n),
        f = Ua(l.baseId, n),
        m = n === l.value,
        c = $(m);
      return (
        q(() => {
          let d = requestAnimationFrame(() => (c.current = !1));
          return () => cancelAnimationFrame(d);
        }, []),
        o(Nn, {
          present: a || m,
          children: ({ present: d }) =>
            o(ze.div, {
              "data-state": m ? "active" : "inactive",
              "data-orientation": l.orientation,
              role: "tabpanel",
              "aria-labelledby": u,
              hidden: !d,
              id: f,
              tabIndex: 0,
              ...i,
              ref: t,
              style: {
                ...e.style,
                animationDuration: c.current ? "0s" : void 0,
              },
              children: d && s,
            }),
        })
      );
    });
  Va.displayName = $a;
  function Ga(e, t) {
    return `${e}-trigger-${t}`;
  }
  function Ua(e, t) {
    return `${e}-content-${t}`;
  }
  var ja = Ha,
    Tn = Ba,
    Rn = qa,
    kn = Va;
  var ke = ja,
    _e = H(({ className: e, ...t }, r) =>
      o(Tn, {
        ref: r,
        className: D(
          "inline-flex h-10 items-center justify-center rounded-md bg-slate-100 p-1 text-slate-500 dark:bg-slate-800 dark:text-slate-400",
          e
        ),
        ...t,
      })
    );
  _e.displayName = Tn.displayName;
  var Y = H(({ className: e, ...t }, r) =>
    o(Rn, {
      ref: r,
      className: D(
        "inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-white transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-950 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-white data-[state=active]:text-slate-950 data-[state=active]:shadow-sm dark:ring-offset-slate-950 dark:focus-visible:ring-slate-300 dark:data-[state=active]:bg-slate-950 dark:data-[state=active]:text-slate-50",
        e
      ),
      ...t,
    })
  );
  Y.displayName = Rn.displayName;
  var X = H(({ className: e, ...t }, r) =>
    o(kn, {
      ref: r,
      className: D(
        "mt-2 ring-offset-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-950 focus-visible:ring-offset-2 dark:ring-offset-slate-950 dark:focus-visible:ring-slate-300",
        e
      ),
      ...t,
    })
  );
  X.displayName = kn.displayName;
  var ft = class extends Map {
    constructor(t, r = Xc) {
      if (
        (super(),
        Object.defineProperties(this, {
          _intern: { value: new Map() },
          _key: { value: r },
        }),
        t != null)
      )
        for (let [n, a] of t) this.set(n, a);
    }
    get(t) {
      return super.get(Wa(this, t));
    }
    has(t) {
      return super.has(Wa(this, t));
    }
    set(t, r) {
      return super.set(Wc(this, t), r);
    }
    delete(t) {
      return super.delete(Yc(this, t));
    }
  };
  function Wa({ _intern: e, _key: t }, r) {
    let n = t(r);
    return e.has(n) ? e.get(n) : r;
  }
  function Wc({ _intern: e, _key: t }, r) {
    let n = t(r);
    return e.has(n) ? e.get(n) : (e.set(n, r), r);
  }
  function Yc({ _intern: e, _key: t }, r) {
    let n = t(r);
    return e.has(n) && ((r = e.get(n)), e.delete(n)), r;
  }
  function Xc(e) {
    return e !== null && typeof e == "object" ? e.valueOf() : e;
  }
  var Qc = { value: () => {} };
  function Xa() {
    for (var e = 0, t = arguments.length, r = {}, n; e < t; ++e) {
      if (!(n = arguments[e] + "") || n in r || /[\s.]/.test(n))
        throw new Error("illegal type: " + n);
      r[n] = [];
    }
    return new hr(r);
  }
  function hr(e) {
    this._ = e;
  }
  function Kc(e, t) {
    return e
      .trim()
      .split(/^|\s+/)
      .map(function (r) {
        var n = "",
          a = r.indexOf(".");
        if (
          (a >= 0 && ((n = r.slice(a + 1)), (r = r.slice(0, a))),
          r && !t.hasOwnProperty(r))
        )
          throw new Error("unknown type: " + r);
        return { type: r, name: n };
      });
  }
  hr.prototype = Xa.prototype = {
    constructor: hr,
    on: function (e, t) {
      var r = this._,
        n = Kc(e + "", r),
        a,
        s = -1,
        i = n.length;
      if (arguments.length < 2) {
        for (; ++s < i; )
          if ((a = (e = n[s]).type) && (a = Jc(r[a], e.name))) return a;
        return;
      }
      if (t != null && typeof t != "function")
        throw new Error("invalid callback: " + t);
      for (; ++s < i; )
        if ((a = (e = n[s]).type)) r[a] = Ya(r[a], e.name, t);
        else if (t == null) for (a in r) r[a] = Ya(r[a], e.name, null);
      return this;
    },
    copy: function () {
      var e = {},
        t = this._;
      for (var r in t) e[r] = t[r].slice();
      return new hr(e);
    },
    call: function (e, t) {
      if ((a = arguments.length - 2) > 0)
        for (var r = new Array(a), n = 0, a, s; n < a; ++n)
          r[n] = arguments[n + 2];
      if (!this._.hasOwnProperty(e)) throw new Error("unknown type: " + e);
      for (s = this._[e], n = 0, a = s.length; n < a; ++n)
        s[n].value.apply(t, r);
    },
    apply: function (e, t, r) {
      if (!this._.hasOwnProperty(e)) throw new Error("unknown type: " + e);
      for (var n = this._[e], a = 0, s = n.length; a < s; ++a)
        n[a].value.apply(t, r);
    },
  };
  function Jc(e, t) {
    for (var r = 0, n = e.length, a; r < n; ++r)
      if ((a = e[r]).name === t) return a.value;
  }
  function Ya(e, t, r) {
    for (var n = 0, a = e.length; n < a; ++n)
      if (e[n].name === t) {
        (e[n] = Qc), (e = e.slice(0, n).concat(e.slice(n + 1)));
        break;
      }
    return r != null && e.push({ name: t, value: r }), e;
  }
  var Sn = Xa;
  var gr = "http://www.w3.org/1999/xhtml",
    En = {
      svg: "http://www.w3.org/2000/svg",
      xhtml: gr,
      xlink: "http://www.w3.org/1999/xlink",
      xml: "http://www.w3.org/XML/1998/namespace",
      xmlns: "http://www.w3.org/2000/xmlns/",
    };
  function Ae(e) {
    var t = (e += ""),
      r = t.indexOf(":");
    return (
      r >= 0 && (t = e.slice(0, r)) !== "xmlns" && (e = e.slice(r + 1)),
      En.hasOwnProperty(t) ? { space: En[t], local: e } : e
    );
  }
  function Zc(e) {
    return function () {
      var t = this.ownerDocument,
        r = this.namespaceURI;
      return r === gr && t.documentElement.namespaceURI === gr
        ? t.createElement(e)
        : t.createElementNS(r, e);
    };
  }
  function eu(e) {
    return function () {
      return this.ownerDocument.createElementNS(e.space, e.local);
    };
  }
  function vr(e) {
    var t = Ae(e);
    return (t.local ? eu : Zc)(t);
  }
  function tu() {}
  function Je(e) {
    return e == null
      ? tu
      : function () {
          return this.querySelector(e);
        };
  }
  function Qa(e) {
    typeof e != "function" && (e = Je(e));
    for (
      var t = this._groups, r = t.length, n = new Array(r), a = 0;
      a < r;
      ++a
    )
      for (
        var s = t[a], i = s.length, l = (n[a] = new Array(i)), u, f, m = 0;
        m < i;
        ++m
      )
        (u = s[m]) &&
          (f = e.call(u, u.__data__, m, s)) &&
          ("__data__" in u && (f.__data__ = u.__data__), (l[m] = f));
    return new te(n, this._parents);
  }
  function Mn(e) {
    return e == null ? [] : Array.isArray(e) ? e : Array.from(e);
  }
  function ru() {
    return [];
  }
  function Mt(e) {
    return e == null
      ? ru
      : function () {
          return this.querySelectorAll(e);
        };
  }
  function nu(e) {
    return function () {
      return Mn(e.apply(this, arguments));
    };
  }
  function Ka(e) {
    typeof e == "function" ? (e = nu(e)) : (e = Mt(e));
    for (var t = this._groups, r = t.length, n = [], a = [], s = 0; s < r; ++s)
      for (var i = t[s], l = i.length, u, f = 0; f < l; ++f)
        (u = i[f]) && (n.push(e.call(u, u.__data__, f, i)), a.push(u));
    return new te(n, a);
  }
  function Pt(e) {
    return function () {
      return this.matches(e);
    };
  }
  function xr(e) {
    return function (t) {
      return t.matches(e);
    };
  }
  var ou = Array.prototype.find;
  function au(e) {
    return function () {
      return ou.call(this.children, e);
    };
  }
  function su() {
    return this.firstElementChild;
  }
  function Ja(e) {
    return this.select(e == null ? su : au(typeof e == "function" ? e : xr(e)));
  }
  var iu = Array.prototype.filter;
  function lu() {
    return Array.from(this.children);
  }
  function cu(e) {
    return function () {
      return iu.call(this.children, e);
    };
  }
  function Za(e) {
    return this.selectAll(
      e == null ? lu : cu(typeof e == "function" ? e : xr(e))
    );
  }
  function es(e) {
    typeof e != "function" && (e = Pt(e));
    for (
      var t = this._groups, r = t.length, n = new Array(r), a = 0;
      a < r;
      ++a
    )
      for (var s = t[a], i = s.length, l = (n[a] = []), u, f = 0; f < i; ++f)
        (u = s[f]) && e.call(u, u.__data__, f, s) && l.push(u);
    return new te(n, this._parents);
  }
  function br(e) {
    return new Array(e.length);
  }
  function ts() {
    return new te(this._enter || this._groups.map(br), this._parents);
  }
  function It(e, t) {
    (this.ownerDocument = e.ownerDocument),
      (this.namespaceURI = e.namespaceURI),
      (this._next = null),
      (this._parent = e),
      (this.__data__ = t);
  }
  It.prototype = {
    constructor: It,
    appendChild: function (e) {
      return this._parent.insertBefore(e, this._next);
    },
    insertBefore: function (e, t) {
      return this._parent.insertBefore(e, t);
    },
    querySelector: function (e) {
      return this._parent.querySelector(e);
    },
    querySelectorAll: function (e) {
      return this._parent.querySelectorAll(e);
    },
  };
  function rs(e) {
    return function () {
      return e;
    };
  }
  function uu(e, t, r, n, a, s) {
    for (var i = 0, l, u = t.length, f = s.length; i < f; ++i)
      (l = t[i]) ? ((l.__data__ = s[i]), (n[i] = l)) : (r[i] = new It(e, s[i]));
    for (; i < u; ++i) (l = t[i]) && (a[i] = l);
  }
  function fu(e, t, r, n, a, s, i) {
    var l,
      u,
      f = new Map(),
      m = t.length,
      c = s.length,
      d = new Array(m),
      p;
    for (l = 0; l < m; ++l)
      (u = t[l]) &&
        ((d[l] = p = i.call(u, u.__data__, l, t) + ""),
        f.has(p) ? (a[l] = u) : f.set(p, u));
    for (l = 0; l < c; ++l)
      (p = i.call(e, s[l], l, s) + ""),
        (u = f.get(p))
          ? ((n[l] = u), (u.__data__ = s[l]), f.delete(p))
          : (r[l] = new It(e, s[l]));
    for (l = 0; l < m; ++l) (u = t[l]) && f.get(d[l]) === u && (a[l] = u);
  }
  function du(e) {
    return e.__data__;
  }
  function ns(e, t) {
    if (!arguments.length) return Array.from(this, du);
    var r = t ? fu : uu,
      n = this._parents,
      a = this._groups;
    typeof e != "function" && (e = rs(e));
    for (
      var s = a.length,
        i = new Array(s),
        l = new Array(s),
        u = new Array(s),
        f = 0;
      f < s;
      ++f
    ) {
      var m = n[f],
        c = a[f],
        d = c.length,
        p = mu(e.call(m, m && m.__data__, f, n)),
        x = p.length,
        b = (l[f] = new Array(x)),
        v = (i[f] = new Array(x)),
        y = (u[f] = new Array(d));
      r(m, c, b, v, y, p, t);
      for (var N = 0, R = 0, M, w; N < x; ++N)
        if ((M = b[N])) {
          for (N >= R && (R = N + 1); !(w = v[R]) && ++R < x; );
          M._next = w || null;
        }
    }
    return (i = new te(i, n)), (i._enter = l), (i._exit = u), i;
  }
  function mu(e) {
    return typeof e == "object" && "length" in e ? e : Array.from(e);
  }
  function os() {
    return new te(this._exit || this._groups.map(br), this._parents);
  }
  function as(e, t, r) {
    var n = this.enter(),
      a = this,
      s = this.exit();
    return (
      typeof e == "function"
        ? ((n = e(n)), n && (n = n.selection()))
        : (n = n.append(e + "")),
      t != null && ((a = t(a)), a && (a = a.selection())),
      r == null ? s.remove() : r(s),
      n && a ? n.merge(a).order() : a
    );
  }
  function ss(e) {
    for (
      var t = e.selection ? e.selection() : e,
        r = this._groups,
        n = t._groups,
        a = r.length,
        s = n.length,
        i = Math.min(a, s),
        l = new Array(a),
        u = 0;
      u < i;
      ++u
    )
      for (
        var f = r[u],
          m = n[u],
          c = f.length,
          d = (l[u] = new Array(c)),
          p,
          x = 0;
        x < c;
        ++x
      )
        (p = f[x] || m[x]) && (d[x] = p);
    for (; u < a; ++u) l[u] = r[u];
    return new te(l, this._parents);
  }
  function is() {
    for (var e = this._groups, t = -1, r = e.length; ++t < r; )
      for (var n = e[t], a = n.length - 1, s = n[a], i; --a >= 0; )
        (i = n[a]) &&
          (s &&
            i.compareDocumentPosition(s) ^ 4 &&
            s.parentNode.insertBefore(i, s),
          (s = i));
    return this;
  }
  function ls(e) {
    e || (e = pu);
    function t(c, d) {
      return c && d ? e(c.__data__, d.__data__) : !c - !d;
    }
    for (
      var r = this._groups, n = r.length, a = new Array(n), s = 0;
      s < n;
      ++s
    ) {
      for (
        var i = r[s], l = i.length, u = (a[s] = new Array(l)), f, m = 0;
        m < l;
        ++m
      )
        (f = i[m]) && (u[m] = f);
      u.sort(t);
    }
    return new te(a, this._parents).order();
  }
  function pu(e, t) {
    return e < t ? -1 : e > t ? 1 : e >= t ? 0 : NaN;
  }
  function cs() {
    var e = arguments[0];
    return (arguments[0] = this), e.apply(null, arguments), this;
  }
  function us() {
    return Array.from(this);
  }
  function fs() {
    for (var e = this._groups, t = 0, r = e.length; t < r; ++t)
      for (var n = e[t], a = 0, s = n.length; a < s; ++a) {
        var i = n[a];
        if (i) return i;
      }
    return null;
  }
  function ds() {
    let e = 0;
    for (let t of this) ++e;
    return e;
  }
  function ms() {
    return !this.node();
  }
  function ps(e) {
    for (var t = this._groups, r = 0, n = t.length; r < n; ++r)
      for (var a = t[r], s = 0, i = a.length, l; s < i; ++s)
        (l = a[s]) && e.call(l, l.__data__, s, a);
    return this;
  }
  function hu(e) {
    return function () {
      this.removeAttribute(e);
    };
  }
  function gu(e) {
    return function () {
      this.removeAttributeNS(e.space, e.local);
    };
  }
  function vu(e, t) {
    return function () {
      this.setAttribute(e, t);
    };
  }
  function xu(e, t) {
    return function () {
      this.setAttributeNS(e.space, e.local, t);
    };
  }
  function bu(e, t) {
    return function () {
      var r = t.apply(this, arguments);
      r == null ? this.removeAttribute(e) : this.setAttribute(e, r);
    };
  }
  function yu(e, t) {
    return function () {
      var r = t.apply(this, arguments);
      r == null
        ? this.removeAttributeNS(e.space, e.local)
        : this.setAttributeNS(e.space, e.local, r);
    };
  }
  function hs(e, t) {
    var r = Ae(e);
    if (arguments.length < 2) {
      var n = this.node();
      return r.local ? n.getAttributeNS(r.space, r.local) : n.getAttribute(r);
    }
    return this.each(
      (t == null
        ? r.local
          ? gu
          : hu
        : typeof t == "function"
        ? r.local
          ? yu
          : bu
        : r.local
        ? xu
        : vu)(r, t)
    );
  }
  function yr(e) {
    return (
      (e.ownerDocument && e.ownerDocument.defaultView) ||
      (e.document && e) ||
      e.defaultView
    );
  }
  function _u(e) {
    return function () {
      this.style.removeProperty(e);
    };
  }
  function wu(e, t, r) {
    return function () {
      this.style.setProperty(e, t, r);
    };
  }
  function Nu(e, t, r) {
    return function () {
      var n = t.apply(this, arguments);
      n == null
        ? this.style.removeProperty(e)
        : this.style.setProperty(e, n, r);
    };
  }
  function gs(e, t, r) {
    return arguments.length > 1
      ? this.each(
          (t == null ? _u : typeof t == "function" ? Nu : wu)(e, t, r ?? "")
        )
      : qe(this.node(), e);
  }
  function qe(e, t) {
    return (
      e.style.getPropertyValue(t) ||
      yr(e).getComputedStyle(e, null).getPropertyValue(t)
    );
  }
  function Cu(e) {
    return function () {
      delete this[e];
    };
  }
  function Tu(e, t) {
    return function () {
      this[e] = t;
    };
  }
  function Ru(e, t) {
    return function () {
      var r = t.apply(this, arguments);
      r == null ? delete this[e] : (this[e] = r);
    };
  }
  function vs(e, t) {
    return arguments.length > 1
      ? this.each((t == null ? Cu : typeof t == "function" ? Ru : Tu)(e, t))
      : this.node()[e];
  }
  function xs(e) {
    return e.trim().split(/^|\s+/);
  }
  function Pn(e) {
    return e.classList || new bs(e);
  }
  function bs(e) {
    (this._node = e), (this._names = xs(e.getAttribute("class") || ""));
  }
  bs.prototype = {
    add: function (e) {
      var t = this._names.indexOf(e);
      t < 0 &&
        (this._names.push(e),
        this._node.setAttribute("class", this._names.join(" ")));
    },
    remove: function (e) {
      var t = this._names.indexOf(e);
      t >= 0 &&
        (this._names.splice(t, 1),
        this._node.setAttribute("class", this._names.join(" ")));
    },
    contains: function (e) {
      return this._names.indexOf(e) >= 0;
    },
  };
  function ys(e, t) {
    for (var r = Pn(e), n = -1, a = t.length; ++n < a; ) r.add(t[n]);
  }
  function _s(e, t) {
    for (var r = Pn(e), n = -1, a = t.length; ++n < a; ) r.remove(t[n]);
  }
  function ku(e) {
    return function () {
      ys(this, e);
    };
  }
  function Su(e) {
    return function () {
      _s(this, e);
    };
  }
  function Eu(e, t) {
    return function () {
      (t.apply(this, arguments) ? ys : _s)(this, e);
    };
  }
  function ws(e, t) {
    var r = xs(e + "");
    if (arguments.length < 2) {
      for (var n = Pn(this.node()), a = -1, s = r.length; ++a < s; )
        if (!n.contains(r[a])) return !1;
      return !0;
    }
    return this.each((typeof t == "function" ? Eu : t ? ku : Su)(r, t));
  }
  function Mu() {
    this.textContent = "";
  }
  function Pu(e) {
    return function () {
      this.textContent = e;
    };
  }
  function Iu(e) {
    return function () {
      var t = e.apply(this, arguments);
      this.textContent = t ?? "";
    };
  }
  function Ns(e) {
    return arguments.length
      ? this.each(e == null ? Mu : (typeof e == "function" ? Iu : Pu)(e))
      : this.node().textContent;
  }
  function Au() {
    this.innerHTML = "";
  }
  function Du(e) {
    return function () {
      this.innerHTML = e;
    };
  }
  function Lu(e) {
    return function () {
      var t = e.apply(this, arguments);
      this.innerHTML = t ?? "";
    };
  }
  function Cs(e) {
    return arguments.length
      ? this.each(e == null ? Au : (typeof e == "function" ? Lu : Du)(e))
      : this.node().innerHTML;
  }
  function Ou() {
    this.nextSibling && this.parentNode.appendChild(this);
  }
  function Ts() {
    return this.each(Ou);
  }
  function Hu() {
    this.previousSibling &&
      this.parentNode.insertBefore(this, this.parentNode.firstChild);
  }
  function Rs() {
    return this.each(Hu);
  }
  function ks(e) {
    var t = typeof e == "function" ? e : vr(e);
    return this.select(function () {
      return this.appendChild(t.apply(this, arguments));
    });
  }
  function Fu() {
    return null;
  }
  function Ss(e, t) {
    var r = typeof e == "function" ? e : vr(e),
      n = t == null ? Fu : typeof t == "function" ? t : Je(t);
    return this.select(function () {
      return this.insertBefore(
        r.apply(this, arguments),
        n.apply(this, arguments) || null
      );
    });
  }
  function Bu() {
    var e = this.parentNode;
    e && e.removeChild(this);
  }
  function Es() {
    return this.each(Bu);
  }
  function zu() {
    var e = this.cloneNode(!1),
      t = this.parentNode;
    return t ? t.insertBefore(e, this.nextSibling) : e;
  }
  function qu() {
    var e = this.cloneNode(!0),
      t = this.parentNode;
    return t ? t.insertBefore(e, this.nextSibling) : e;
  }
  function Ms(e) {
    return this.select(e ? qu : zu);
  }
  function Ps(e) {
    return arguments.length
      ? this.property("__data__", e)
      : this.node().__data__;
  }
  function $u(e) {
    return function (t) {
      e.call(this, t, this.__data__);
    };
  }
  function Vu(e) {
    return e
      .trim()
      .split(/^|\s+/)
      .map(function (t) {
        var r = "",
          n = t.indexOf(".");
        return (
          n >= 0 && ((r = t.slice(n + 1)), (t = t.slice(0, n))),
          { type: t, name: r }
        );
      });
  }
  function Gu(e) {
    return function () {
      var t = this.__on;
      if (t) {
        for (var r = 0, n = -1, a = t.length, s; r < a; ++r)
          (s = t[r]),
            (!e.type || s.type === e.type) && s.name === e.name
              ? this.removeEventListener(s.type, s.listener, s.options)
              : (t[++n] = s);
        ++n ? (t.length = n) : delete this.__on;
      }
    };
  }
  function Uu(e, t, r) {
    return function () {
      var n = this.__on,
        a,
        s = $u(t);
      if (n) {
        for (var i = 0, l = n.length; i < l; ++i)
          if ((a = n[i]).type === e.type && a.name === e.name) {
            this.removeEventListener(a.type, a.listener, a.options),
              this.addEventListener(a.type, (a.listener = s), (a.options = r)),
              (a.value = t);
            return;
          }
      }
      this.addEventListener(e.type, s, r),
        (a = { type: e.type, name: e.name, value: t, listener: s, options: r }),
        n ? n.push(a) : (this.__on = [a]);
    };
  }
  function Is(e, t, r) {
    var n = Vu(e + ""),
      a,
      s = n.length,
      i;
    if (arguments.length < 2) {
      var l = this.node().__on;
      if (l) {
        for (var u = 0, f = l.length, m; u < f; ++u)
          for (a = 0, m = l[u]; a < s; ++a)
            if ((i = n[a]).type === m.type && i.name === m.name) return m.value;
      }
      return;
    }
    for (l = t ? Uu : Gu, a = 0; a < s; ++a) this.each(l(n[a], t, r));
    return this;
  }
  function As(e, t, r) {
    var n = yr(e),
      a = n.CustomEvent;
    typeof a == "function"
      ? (a = new a(t, r))
      : ((a = n.document.createEvent("Event")),
        r
          ? (a.initEvent(t, r.bubbles, r.cancelable), (a.detail = r.detail))
          : a.initEvent(t, !1, !1)),
      e.dispatchEvent(a);
  }
  function ju(e, t) {
    return function () {
      return As(this, e, t);
    };
  }
  function Wu(e, t) {
    return function () {
      return As(this, e, t.apply(this, arguments));
    };
  }
  function Ds(e, t) {
    return this.each((typeof t == "function" ? Wu : ju)(e, t));
  }
  function* Ls() {
    for (var e = this._groups, t = 0, r = e.length; t < r; ++t)
      for (var n = e[t], a = 0, s = n.length, i; a < s; ++a)
        (i = n[a]) && (yield i);
  }
  var In = [null];
  function te(e, t) {
    (this._groups = e), (this._parents = t);
  }
  function Os() {
    return new te([[document.documentElement]], In);
  }
  function Yu() {
    return this;
  }
  te.prototype = Os.prototype = {
    constructor: te,
    select: Qa,
    selectAll: Ka,
    selectChild: Ja,
    selectChildren: Za,
    filter: es,
    data: ns,
    enter: ts,
    exit: os,
    join: as,
    merge: ss,
    selection: Yu,
    order: is,
    sort: ls,
    call: cs,
    nodes: us,
    node: fs,
    size: ds,
    empty: ms,
    each: ps,
    attr: hs,
    style: gs,
    property: vs,
    classed: ws,
    text: Ns,
    html: Cs,
    raise: Ts,
    lower: Rs,
    append: ks,
    insert: Ss,
    remove: Es,
    clone: Ms,
    datum: Ps,
    on: Is,
    dispatch: Ds,
    [Symbol.iterator]: Ls,
  };
  var De = Os;
  function An(e) {
    return typeof e == "string"
      ? new te([[document.querySelector(e)]], [document.documentElement])
      : new te([[e]], In);
  }
  function _r(e, t, r) {
    (e.prototype = t.prototype = r), (r.constructor = e);
  }
  function Dn(e, t) {
    var r = Object.create(e.prototype);
    for (var n in t) r[n] = t[n];
    return r;
  }
  function Lt() {}
  var At = 0.7,
    Cr = 1 / At,
    dt = "\\s*([+-]?\\d+)\\s*",
    Dt = "\\s*([+-]?(?:\\d*\\.)?\\d+(?:[eE][+-]?\\d+)?)\\s*",
    Se = "\\s*([+-]?(?:\\d*\\.)?\\d+(?:[eE][+-]?\\d+)?)%\\s*",
    Xu = /^#([0-9a-f]{3,8})$/,
    Qu = new RegExp(`^rgb\\(${dt},${dt},${dt}\\)$`),
    Ku = new RegExp(`^rgb\\(${Se},${Se},${Se}\\)$`),
    Ju = new RegExp(`^rgba\\(${dt},${dt},${dt},${Dt}\\)$`),
    Zu = new RegExp(`^rgba\\(${Se},${Se},${Se},${Dt}\\)$`),
    ef = new RegExp(`^hsl\\(${Dt},${Se},${Se}\\)$`),
    tf = new RegExp(`^hsla\\(${Dt},${Se},${Se},${Dt}\\)$`),
    Hs = {
      aliceblue: 15792383,
      antiquewhite: 16444375,
      aqua: 65535,
      aquamarine: 8388564,
      azure: 15794175,
      beige: 16119260,
      bisque: 16770244,
      black: 0,
      blanchedalmond: 16772045,
      blue: 255,
      blueviolet: 9055202,
      brown: 10824234,
      burlywood: 14596231,
      cadetblue: 6266528,
      chartreuse: 8388352,
      chocolate: 13789470,
      coral: 16744272,
      cornflowerblue: 6591981,
      cornsilk: 16775388,
      crimson: 14423100,
      cyan: 65535,
      darkblue: 139,
      darkcyan: 35723,
      darkgoldenrod: 12092939,
      darkgray: 11119017,
      darkgreen: 25600,
      darkgrey: 11119017,
      darkkhaki: 12433259,
      darkmagenta: 9109643,
      darkolivegreen: 5597999,
      darkorange: 16747520,
      darkorchid: 10040012,
      darkred: 9109504,
      darksalmon: 15308410,
      darkseagreen: 9419919,
      darkslateblue: 4734347,
      darkslategray: 3100495,
      darkslategrey: 3100495,
      darkturquoise: 52945,
      darkviolet: 9699539,
      deeppink: 16716947,
      deepskyblue: 49151,
      dimgray: 6908265,
      dimgrey: 6908265,
      dodgerblue: 2003199,
      firebrick: 11674146,
      floralwhite: 16775920,
      forestgreen: 2263842,
      fuchsia: 16711935,
      gainsboro: 14474460,
      ghostwhite: 16316671,
      gold: 16766720,
      goldenrod: 14329120,
      gray: 8421504,
      green: 32768,
      greenyellow: 11403055,
      grey: 8421504,
      honeydew: 15794160,
      hotpink: 16738740,
      indianred: 13458524,
      indigo: 4915330,
      ivory: 16777200,
      khaki: 15787660,
      lavender: 15132410,
      lavenderblush: 16773365,
      lawngreen: 8190976,
      lemonchiffon: 16775885,
      lightblue: 11393254,
      lightcoral: 15761536,
      lightcyan: 14745599,
      lightgoldenrodyellow: 16448210,
      lightgray: 13882323,
      lightgreen: 9498256,
      lightgrey: 13882323,
      lightpink: 16758465,
      lightsalmon: 16752762,
      lightseagreen: 2142890,
      lightskyblue: 8900346,
      lightslategray: 7833753,
      lightslategrey: 7833753,
      lightsteelblue: 11584734,
      lightyellow: 16777184,
      lime: 65280,
      limegreen: 3329330,
      linen: 16445670,
      magenta: 16711935,
      maroon: 8388608,
      mediumaquamarine: 6737322,
      mediumblue: 205,
      mediumorchid: 12211667,
      mediumpurple: 9662683,
      mediumseagreen: 3978097,
      mediumslateblue: 8087790,
      mediumspringgreen: 64154,
      mediumturquoise: 4772300,
      mediumvioletred: 13047173,
      midnightblue: 1644912,
      mintcream: 16121850,
      mistyrose: 16770273,
      moccasin: 16770229,
      navajowhite: 16768685,
      navy: 128,
      oldlace: 16643558,
      olive: 8421376,
      olivedrab: 7048739,
      orange: 16753920,
      orangered: 16729344,
      orchid: 14315734,
      palegoldenrod: 15657130,
      palegreen: 10025880,
      paleturquoise: 11529966,
      palevioletred: 14381203,
      papayawhip: 16773077,
      peachpuff: 16767673,
      peru: 13468991,
      pink: 16761035,
      plum: 14524637,
      powderblue: 11591910,
      purple: 8388736,
      rebeccapurple: 6697881,
      red: 16711680,
      rosybrown: 12357519,
      royalblue: 4286945,
      saddlebrown: 9127187,
      salmon: 16416882,
      sandybrown: 16032864,
      seagreen: 3050327,
      seashell: 16774638,
      sienna: 10506797,
      silver: 12632256,
      skyblue: 8900331,
      slateblue: 6970061,
      slategray: 7372944,
      slategrey: 7372944,
      snow: 16775930,
      springgreen: 65407,
      steelblue: 4620980,
      tan: 13808780,
      teal: 32896,
      thistle: 14204888,
      tomato: 16737095,
      turquoise: 4251856,
      violet: 15631086,
      wheat: 16113331,
      white: 16777215,
      whitesmoke: 16119285,
      yellow: 16776960,
      yellowgreen: 10145074,
    };
  _r(Lt, $e, {
    copy(e) {
      return Object.assign(new this.constructor(), this, e);
    },
    displayable() {
      return this.rgb().displayable();
    },
    hex: Fs,
    formatHex: Fs,
    formatHex8: rf,
    formatHsl: nf,
    formatRgb: Bs,
    toString: Bs,
  });
  function Fs() {
    return this.rgb().formatHex();
  }
  function rf() {
    return this.rgb().formatHex8();
  }
  function nf() {
    return Us(this).formatHsl();
  }
  function Bs() {
    return this.rgb().formatRgb();
  }
  function $e(e) {
    var t, r;
    return (
      (e = (e + "").trim().toLowerCase()),
      (t = Xu.exec(e))
        ? ((r = t[1].length),
          (t = parseInt(t[1], 16)),
          r === 6
            ? zs(t)
            : r === 3
            ? new de(
                ((t >> 8) & 15) | ((t >> 4) & 240),
                ((t >> 4) & 15) | (t & 240),
                ((t & 15) << 4) | (t & 15),
                1
              )
            : r === 8
            ? wr(
                (t >> 24) & 255,
                (t >> 16) & 255,
                (t >> 8) & 255,
                (t & 255) / 255
              )
            : r === 4
            ? wr(
                ((t >> 12) & 15) | ((t >> 8) & 240),
                ((t >> 8) & 15) | ((t >> 4) & 240),
                ((t >> 4) & 15) | (t & 240),
                (((t & 15) << 4) | (t & 15)) / 255
              )
            : null)
        : (t = Qu.exec(e))
        ? new de(t[1], t[2], t[3], 1)
        : (t = Ku.exec(e))
        ? new de((t[1] * 255) / 100, (t[2] * 255) / 100, (t[3] * 255) / 100, 1)
        : (t = Ju.exec(e))
        ? wr(t[1], t[2], t[3], t[4])
        : (t = Zu.exec(e))
        ? wr((t[1] * 255) / 100, (t[2] * 255) / 100, (t[3] * 255) / 100, t[4])
        : (t = ef.exec(e))
        ? Vs(t[1], t[2] / 100, t[3] / 100, 1)
        : (t = tf.exec(e))
        ? Vs(t[1], t[2] / 100, t[3] / 100, t[4])
        : Hs.hasOwnProperty(e)
        ? zs(Hs[e])
        : e === "transparent"
        ? new de(NaN, NaN, NaN, 0)
        : null
    );
  }
  function zs(e) {
    return new de((e >> 16) & 255, (e >> 8) & 255, e & 255, 1);
  }
  function wr(e, t, r, n) {
    return n <= 0 && (e = t = r = NaN), new de(e, t, r, n);
  }
  function of(e) {
    return (
      e instanceof Lt || (e = $e(e)),
      e ? ((e = e.rgb()), new de(e.r, e.g, e.b, e.opacity)) : new de()
    );
  }
  function mt(e, t, r, n) {
    return arguments.length === 1 ? of(e) : new de(e, t, r, n ?? 1);
  }
  function de(e, t, r, n) {
    (this.r = +e), (this.g = +t), (this.b = +r), (this.opacity = +n);
  }
  _r(
    de,
    mt,
    Dn(Lt, {
      brighter(e) {
        return (
          (e = e == null ? Cr : Math.pow(Cr, e)),
          new de(this.r * e, this.g * e, this.b * e, this.opacity)
        );
      },
      darker(e) {
        return (
          (e = e == null ? At : Math.pow(At, e)),
          new de(this.r * e, this.g * e, this.b * e, this.opacity)
        );
      },
      rgb() {
        return this;
      },
      clamp() {
        return new de(et(this.r), et(this.g), et(this.b), Tr(this.opacity));
      },
      displayable() {
        return (
          -0.5 <= this.r &&
          this.r < 255.5 &&
          -0.5 <= this.g &&
          this.g < 255.5 &&
          -0.5 <= this.b &&
          this.b < 255.5 &&
          0 <= this.opacity &&
          this.opacity <= 1
        );
      },
      hex: qs,
      formatHex: qs,
      formatHex8: af,
      formatRgb: $s,
      toString: $s,
    })
  );
  function qs() {
    return `#${Ze(this.r)}${Ze(this.g)}${Ze(this.b)}`;
  }
  function af() {
    return `#${Ze(this.r)}${Ze(this.g)}${Ze(this.b)}${Ze(
      (isNaN(this.opacity) ? 1 : this.opacity) * 255
    )}`;
  }
  function $s() {
    let e = Tr(this.opacity);
    return `${e === 1 ? "rgb(" : "rgba("}${et(this.r)}, ${et(this.g)}, ${et(
      this.b
    )}${e === 1 ? ")" : `, ${e})`}`;
  }
  function Tr(e) {
    return isNaN(e) ? 1 : Math.max(0, Math.min(1, e));
  }
  function et(e) {
    return Math.max(0, Math.min(255, Math.round(e) || 0));
  }
  function Ze(e) {
    return (e = et(e)), (e < 16 ? "0" : "") + e.toString(16);
  }
  function Vs(e, t, r, n) {
    return (
      n <= 0
        ? (e = t = r = NaN)
        : r <= 0 || r >= 1
        ? (e = t = NaN)
        : t <= 0 && (e = NaN),
      new we(e, t, r, n)
    );
  }
  function Us(e) {
    if (e instanceof we) return new we(e.h, e.s, e.l, e.opacity);
    if ((e instanceof Lt || (e = $e(e)), !e)) return new we();
    if (e instanceof we) return e;
    e = e.rgb();
    var t = e.r / 255,
      r = e.g / 255,
      n = e.b / 255,
      a = Math.min(t, r, n),
      s = Math.max(t, r, n),
      i = NaN,
      l = s - a,
      u = (s + a) / 2;
    return (
      l
        ? (t === s
            ? (i = (r - n) / l + (r < n) * 6)
            : r === s
            ? (i = (n - t) / l + 2)
            : (i = (t - r) / l + 4),
          (l /= u < 0.5 ? s + a : 2 - s - a),
          (i *= 60))
        : (l = u > 0 && u < 1 ? 0 : i),
      new we(i, l, u, e.opacity)
    );
  }
  function js(e, t, r, n) {
    return arguments.length === 1 ? Us(e) : new we(e, t, r, n ?? 1);
  }
  function we(e, t, r, n) {
    (this.h = +e), (this.s = +t), (this.l = +r), (this.opacity = +n);
  }
  _r(
    we,
    js,
    Dn(Lt, {
      brighter(e) {
        return (
          (e = e == null ? Cr : Math.pow(Cr, e)),
          new we(this.h, this.s, this.l * e, this.opacity)
        );
      },
      darker(e) {
        return (
          (e = e == null ? At : Math.pow(At, e)),
          new we(this.h, this.s, this.l * e, this.opacity)
        );
      },
      rgb() {
        var e = (this.h % 360) + (this.h < 0) * 360,
          t = isNaN(e) || isNaN(this.s) ? 0 : this.s,
          r = this.l,
          n = r + (r < 0.5 ? r : 1 - r) * t,
          a = 2 * r - n;
        return new de(
          Ln(e >= 240 ? e - 240 : e + 120, a, n),
          Ln(e, a, n),
          Ln(e < 120 ? e + 240 : e - 120, a, n),
          this.opacity
        );
      },
      clamp() {
        return new we(Gs(this.h), Nr(this.s), Nr(this.l), Tr(this.opacity));
      },
      displayable() {
        return (
          ((0 <= this.s && this.s <= 1) || isNaN(this.s)) &&
          0 <= this.l &&
          this.l <= 1 &&
          0 <= this.opacity &&
          this.opacity <= 1
        );
      },
      formatHsl() {
        let e = Tr(this.opacity);
        return `${e === 1 ? "hsl(" : "hsla("}${Gs(this.h)}, ${
          Nr(this.s) * 100
        }%, ${Nr(this.l) * 100}%${e === 1 ? ")" : `, ${e})`}`;
      },
    })
  );
  function Gs(e) {
    return (e = (e || 0) % 360), e < 0 ? e + 360 : e;
  }
  function Nr(e) {
    return Math.max(0, Math.min(1, e || 0));
  }
  function Ln(e, t, r) {
    return (
      (e < 60
        ? t + ((r - t) * e) / 60
        : e < 180
        ? r
        : e < 240
        ? t + ((r - t) * (240 - e)) / 60
        : t) * 255
    );
  }
  function On(e, t, r, n, a) {
    var s = e * e,
      i = s * e;
    return (
      ((1 - 3 * e + 3 * s - i) * t +
        (4 - 6 * s + 3 * i) * r +
        (1 + 3 * e + 3 * s - 3 * i) * n +
        i * a) /
      6
    );
  }
  function Ws(e) {
    var t = e.length - 1;
    return function (r) {
      var n = r <= 0 ? (r = 0) : r >= 1 ? ((r = 1), t - 1) : Math.floor(r * t),
        a = e[n],
        s = e[n + 1],
        i = n > 0 ? e[n - 1] : 2 * a - s,
        l = n < t - 1 ? e[n + 2] : 2 * s - a;
      return On((r - n / t) * t, i, a, s, l);
    };
  }
  function Ys(e) {
    var t = e.length;
    return function (r) {
      var n = Math.floor(((r %= 1) < 0 ? ++r : r) * t),
        a = e[(n + t - 1) % t],
        s = e[n % t],
        i = e[(n + 1) % t],
        l = e[(n + 2) % t];
      return On((r - n / t) * t, a, s, i, l);
    };
  }
  var Hn = (e) => () => e;
  function sf(e, t) {
    return function (r) {
      return e + r * t;
    };
  }
  function lf(e, t, r) {
    return (
      (e = Math.pow(e, r)),
      (t = Math.pow(t, r) - e),
      (r = 1 / r),
      function (n) {
        return Math.pow(e + n * t, r);
      }
    );
  }
  function Xs(e) {
    return (e = +e) == 1
      ? Rr
      : function (t, r) {
          return r - t ? lf(t, r, e) : Hn(isNaN(t) ? r : t);
        };
  }
  function Rr(e, t) {
    var r = t - e;
    return r ? sf(e, r) : Hn(isNaN(e) ? t : e);
  }
  var kr = (function e(t) {
    var r = Xs(t);
    function n(a, s) {
      var i = r((a = mt(a)).r, (s = mt(s)).r),
        l = r(a.g, s.g),
        u = r(a.b, s.b),
        f = Rr(a.opacity, s.opacity);
      return function (m) {
        return (
          (a.r = i(m)), (a.g = l(m)), (a.b = u(m)), (a.opacity = f(m)), a + ""
        );
      };
    }
    return (n.gamma = e), n;
  })(1);
  function Qs(e) {
    return function (t) {
      var r = t.length,
        n = new Array(r),
        a = new Array(r),
        s = new Array(r),
        i,
        l;
      for (i = 0; i < r; ++i)
        (l = mt(t[i])), (n[i] = l.r || 0), (a[i] = l.g || 0), (s[i] = l.b || 0);
      return (
        (n = e(n)),
        (a = e(a)),
        (s = e(s)),
        (l.opacity = 1),
        function (u) {
          return (l.r = n(u)), (l.g = a(u)), (l.b = s(u)), l + "";
        }
      );
    };
  }
  var cf = Qs(Ws),
    uf = Qs(Ys);
  function xe(e, t) {
    return (
      (e = +e),
      (t = +t),
      function (r) {
        return e * (1 - r) + t * r;
      }
    );
  }
  var Bn = /[-+]?(?:\d+\.?\d*|\.?\d+)(?:[eE][-+]?\d+)?/g,
    Fn = new RegExp(Bn.source, "g");
  function ff(e) {
    return function () {
      return e;
    };
  }
  function df(e) {
    return function (t) {
      return e(t) + "";
    };
  }
  function zn(e, t) {
    var r = (Bn.lastIndex = Fn.lastIndex = 0),
      n,
      a,
      s,
      i = -1,
      l = [],
      u = [];
    for (e = e + "", t = t + ""; (n = Bn.exec(e)) && (a = Fn.exec(t)); )
      (s = a.index) > r &&
        ((s = t.slice(r, s)), l[i] ? (l[i] += s) : (l[++i] = s)),
        (n = n[0]) === (a = a[0])
          ? l[i]
            ? (l[i] += a)
            : (l[++i] = a)
          : ((l[++i] = null), u.push({ i, x: xe(n, a) })),
        (r = Fn.lastIndex);
    return (
      r < t.length && ((s = t.slice(r)), l[i] ? (l[i] += s) : (l[++i] = s)),
      l.length < 2
        ? u[0]
          ? df(u[0].x)
          : ff(t)
        : ((t = u.length),
          function (f) {
            for (var m = 0, c; m < t; ++m) l[(c = u[m]).i] = c.x(f);
            return l.join("");
          })
    );
  }
  var Ks = 180 / Math.PI,
    Sr = {
      translateX: 0,
      translateY: 0,
      rotate: 0,
      skewX: 0,
      scaleX: 1,
      scaleY: 1,
    };
  function qn(e, t, r, n, a, s) {
    var i, l, u;
    return (
      (i = Math.sqrt(e * e + t * t)) && ((e /= i), (t /= i)),
      (u = e * r + t * n) && ((r -= e * u), (n -= t * u)),
      (l = Math.sqrt(r * r + n * n)) && ((r /= l), (n /= l), (u /= l)),
      e * n < t * r && ((e = -e), (t = -t), (u = -u), (i = -i)),
      {
        translateX: a,
        translateY: s,
        rotate: Math.atan2(t, e) * Ks,
        skewX: Math.atan(u) * Ks,
        scaleX: i,
        scaleY: l,
      }
    );
  }
  var Er;
  function Js(e) {
    let t = new (typeof DOMMatrix == "function" ? DOMMatrix : WebKitCSSMatrix)(
      e + ""
    );
    return t.isIdentity ? Sr : qn(t.a, t.b, t.c, t.d, t.e, t.f);
  }
  function Zs(e) {
    return e == null
      ? Sr
      : (Er ||
          (Er = document.createElementNS("http://www.w3.org/2000/svg", "g")),
        Er.setAttribute("transform", e),
        (e = Er.transform.baseVal.consolidate())
          ? ((e = e.matrix), qn(e.a, e.b, e.c, e.d, e.e, e.f))
          : Sr);
  }
  function ei(e, t, r, n) {
    function a(f) {
      return f.length ? f.pop() + " " : "";
    }
    function s(f, m, c, d, p, x) {
      if (f !== c || m !== d) {
        var b = p.push("translate(", null, t, null, r);
        x.push({ i: b - 4, x: xe(f, c) }, { i: b - 2, x: xe(m, d) });
      } else (c || d) && p.push("translate(" + c + t + d + r);
    }
    function i(f, m, c, d) {
      f !== m
        ? (f - m > 180 ? (m += 360) : m - f > 180 && (f += 360),
          d.push({ i: c.push(a(c) + "rotate(", null, n) - 2, x: xe(f, m) }))
        : m && c.push(a(c) + "rotate(" + m + n);
    }
    function l(f, m, c, d) {
      f !== m
        ? d.push({ i: c.push(a(c) + "skewX(", null, n) - 2, x: xe(f, m) })
        : m && c.push(a(c) + "skewX(" + m + n);
    }
    function u(f, m, c, d, p, x) {
      if (f !== c || m !== d) {
        var b = p.push(a(p) + "scale(", null, ",", null, ")");
        x.push({ i: b - 4, x: xe(f, c) }, { i: b - 2, x: xe(m, d) });
      } else
        (c !== 1 || d !== 1) && p.push(a(p) + "scale(" + c + "," + d + ")");
    }
    return function (f, m) {
      var c = [],
        d = [];
      return (
        (f = e(f)),
        (m = e(m)),
        s(f.translateX, f.translateY, m.translateX, m.translateY, c, d),
        i(f.rotate, m.rotate, c, d),
        l(f.skewX, m.skewX, c, d),
        u(f.scaleX, f.scaleY, m.scaleX, m.scaleY, c, d),
        (f = m = null),
        function (p) {
          for (var x = -1, b = d.length, v; ++x < b; ) c[(v = d[x]).i] = v.x(p);
          return c.join("");
        }
      );
    };
  }
  var $n = ei(Js, "px, ", "px)", "deg)"),
    Vn = ei(Zs, ", ", ")", ")");
  var pt = 0,
    Ht = 0,
    Ot = 0,
    ri = 1e3,
    Mr,
    Ft,
    Pr = 0,
    tt = 0,
    Ir = 0,
    Bt = typeof performance == "object" && performance.now ? performance : Date,
    ni =
      typeof window == "object" && window.requestAnimationFrame
        ? window.requestAnimationFrame.bind(window)
        : function (e) {
            setTimeout(e, 17);
          };
  function qt() {
    return tt || (ni(mf), (tt = Bt.now() + Ir));
  }
  function mf() {
    tt = 0;
  }
  function zt() {
    this._call = this._time = this._next = null;
  }
  zt.prototype = Ar.prototype = {
    constructor: zt,
    restart: function (e, t, r) {
      if (typeof e != "function")
        throw new TypeError("callback is not a function");
      (r = (r == null ? qt() : +r) + (t == null ? 0 : +t)),
        !this._next &&
          Ft !== this &&
          (Ft ? (Ft._next = this) : (Mr = this), (Ft = this)),
        (this._call = e),
        (this._time = r),
        Gn();
    },
    stop: function () {
      this._call && ((this._call = null), (this._time = 1 / 0), Gn());
    },
  };
  function Ar(e, t, r) {
    var n = new zt();
    return n.restart(e, t, r), n;
  }
  function oi() {
    qt(), ++pt;
    for (var e = Mr, t; e; )
      (t = tt - e._time) >= 0 && e._call.call(void 0, t), (e = e._next);
    --pt;
  }
  function ti() {
    (tt = (Pr = Bt.now()) + Ir), (pt = Ht = 0);
    try {
      oi();
    } finally {
      (pt = 0), hf(), (tt = 0);
    }
  }
  function pf() {
    var e = Bt.now(),
      t = e - Pr;
    t > ri && ((Ir -= t), (Pr = e));
  }
  function hf() {
    for (var e, t = Mr, r, n = 1 / 0; t; )
      t._call
        ? (n > t._time && (n = t._time), (e = t), (t = t._next))
        : ((r = t._next), (t._next = null), (t = e ? (e._next = r) : (Mr = r)));
    (Ft = e), Gn(n);
  }
  function Gn(e) {
    if (!pt) {
      Ht && (Ht = clearTimeout(Ht));
      var t = e - tt;
      t > 24
        ? (e < 1 / 0 && (Ht = setTimeout(ti, e - Bt.now() - Ir)),
          Ot && (Ot = clearInterval(Ot)))
        : (Ot || ((Pr = Bt.now()), (Ot = setInterval(pf, ri))),
          (pt = 1),
          ni(ti));
    }
  }
  function Dr(e, t, r) {
    var n = new zt();
    return (
      (t = t == null ? 0 : +t),
      n.restart(
        (a) => {
          n.stop(), e(a + t);
        },
        t,
        r
      ),
      n
    );
  }
  var gf = Sn("start", "end", "cancel", "interrupt"),
    vf = [],
    ii = 0,
    ai = 1,
    Or = 2,
    Lr = 3,
    si = 4,
    Hr = 5,
    $t = 6;
  function Ve(e, t, r, n, a, s) {
    var i = e.__transition;
    if (!i) e.__transition = {};
    else if (r in i) return;
    xf(e, r, {
      name: t,
      index: n,
      group: a,
      on: gf,
      tween: vf,
      time: s.time,
      delay: s.delay,
      duration: s.duration,
      ease: s.ease,
      timer: null,
      state: ii,
    });
  }
  function Vt(e, t) {
    var r = ne(e, t);
    if (r.state > ii) throw new Error("too late; already scheduled");
    return r;
  }
  function se(e, t) {
    var r = ne(e, t);
    if (r.state > Lr) throw new Error("too late; already running");
    return r;
  }
  function ne(e, t) {
    var r = e.__transition;
    if (!r || !(r = r[t])) throw new Error("transition not found");
    return r;
  }
  function xf(e, t, r) {
    var n = e.__transition,
      a;
    (n[t] = r), (r.timer = Ar(s, 0, r.time));
    function s(f) {
      (r.state = ai),
        r.timer.restart(i, r.delay, r.time),
        r.delay <= f && i(f - r.delay);
    }
    function i(f) {
      var m, c, d, p;
      if (r.state !== ai) return u();
      for (m in n)
        if (((p = n[m]), p.name === r.name)) {
          if (p.state === Lr) return Dr(i);
          p.state === si
            ? ((p.state = $t),
              p.timer.stop(),
              p.on.call("interrupt", e, e.__data__, p.index, p.group),
              delete n[m])
            : +m < t &&
              ((p.state = $t),
              p.timer.stop(),
              p.on.call("cancel", e, e.__data__, p.index, p.group),
              delete n[m]);
        }
      if (
        (Dr(function () {
          r.state === Lr &&
            ((r.state = si), r.timer.restart(l, r.delay, r.time), l(f));
        }),
        (r.state = Or),
        r.on.call("start", e, e.__data__, r.index, r.group),
        r.state === Or)
      ) {
        for (
          r.state = Lr, a = new Array((d = r.tween.length)), m = 0, c = -1;
          m < d;
          ++m
        )
          (p = r.tween[m].value.call(e, e.__data__, r.index, r.group)) &&
            (a[++c] = p);
        a.length = c + 1;
      }
    }
    function l(f) {
      for (
        var m =
            f < r.duration
              ? r.ease.call(null, f / r.duration)
              : (r.timer.restart(u), (r.state = Hr), 1),
          c = -1,
          d = a.length;
        ++c < d;

      )
        a[c].call(e, m);
      r.state === Hr &&
        (r.on.call("end", e, e.__data__, r.index, r.group), u());
    }
    function u() {
      (r.state = $t), r.timer.stop(), delete n[t];
      for (var f in n) return;
      delete e.__transition;
    }
  }
  function Fr(e, t) {
    var r = e.__transition,
      n,
      a,
      s = !0,
      i;
    if (r) {
      t = t == null ? null : t + "";
      for (i in r) {
        if ((n = r[i]).name !== t) {
          s = !1;
          continue;
        }
        (a = n.state > Or && n.state < Hr),
          (n.state = $t),
          n.timer.stop(),
          n.on.call(
            a ? "interrupt" : "cancel",
            e,
            e.__data__,
            n.index,
            n.group
          ),
          delete r[i];
      }
      s && delete e.__transition;
    }
  }
  function li(e) {
    return this.each(function () {
      Fr(this, e);
    });
  }
  function bf(e, t) {
    var r, n;
    return function () {
      var a = se(this, e),
        s = a.tween;
      if (s !== r) {
        n = r = s;
        for (var i = 0, l = n.length; i < l; ++i)
          if (n[i].name === t) {
            (n = n.slice()), n.splice(i, 1);
            break;
          }
      }
      a.tween = n;
    };
  }
  function yf(e, t, r) {
    var n, a;
    if (typeof r != "function") throw new Error();
    return function () {
      var s = se(this, e),
        i = s.tween;
      if (i !== n) {
        a = (n = i).slice();
        for (var l = { name: t, value: r }, u = 0, f = a.length; u < f; ++u)
          if (a[u].name === t) {
            a[u] = l;
            break;
          }
        u === f && a.push(l);
      }
      s.tween = a;
    };
  }
  function ci(e, t) {
    var r = this._id;
    if (((e += ""), arguments.length < 2)) {
      for (var n = ne(this.node(), r).tween, a = 0, s = n.length, i; a < s; ++a)
        if ((i = n[a]).name === e) return i.value;
      return null;
    }
    return this.each((t == null ? bf : yf)(r, e, t));
  }
  function ht(e, t, r) {
    var n = e._id;
    return (
      e.each(function () {
        var a = se(this, n);
        (a.value || (a.value = {}))[t] = r.apply(this, arguments);
      }),
      function (a) {
        return ne(a, n).value[t];
      }
    );
  }
  function Br(e, t) {
    var r;
    return (
      typeof t == "number"
        ? xe
        : t instanceof $e
        ? kr
        : (r = $e(t))
        ? ((t = r), kr)
        : zn
    )(e, t);
  }
  function _f(e) {
    return function () {
      this.removeAttribute(e);
    };
  }
  function wf(e) {
    return function () {
      this.removeAttributeNS(e.space, e.local);
    };
  }
  function Nf(e, t, r) {
    var n,
      a = r + "",
      s;
    return function () {
      var i = this.getAttribute(e);
      return i === a ? null : i === n ? s : (s = t((n = i), r));
    };
  }
  function Cf(e, t, r) {
    var n,
      a = r + "",
      s;
    return function () {
      var i = this.getAttributeNS(e.space, e.local);
      return i === a ? null : i === n ? s : (s = t((n = i), r));
    };
  }
  function Tf(e, t, r) {
    var n, a, s;
    return function () {
      var i,
        l = r(this),
        u;
      return l == null
        ? void this.removeAttribute(e)
        : ((i = this.getAttribute(e)),
          (u = l + ""),
          i === u
            ? null
            : i === n && u === a
            ? s
            : ((a = u), (s = t((n = i), l))));
    };
  }
  function Rf(e, t, r) {
    var n, a, s;
    return function () {
      var i,
        l = r(this),
        u;
      return l == null
        ? void this.removeAttributeNS(e.space, e.local)
        : ((i = this.getAttributeNS(e.space, e.local)),
          (u = l + ""),
          i === u
            ? null
            : i === n && u === a
            ? s
            : ((a = u), (s = t((n = i), l))));
    };
  }
  function ui(e, t) {
    var r = Ae(e),
      n = r === "transform" ? Vn : Br;
    return this.attrTween(
      e,
      typeof t == "function"
        ? (r.local ? Rf : Tf)(r, n, ht(this, "attr." + e, t))
        : t == null
        ? (r.local ? wf : _f)(r)
        : (r.local ? Cf : Nf)(r, n, t)
    );
  }
  function kf(e, t) {
    return function (r) {
      this.setAttribute(e, t.call(this, r));
    };
  }
  function Sf(e, t) {
    return function (r) {
      this.setAttributeNS(e.space, e.local, t.call(this, r));
    };
  }
  function Ef(e, t) {
    var r, n;
    function a() {
      var s = t.apply(this, arguments);
      return s !== n && (r = (n = s) && Sf(e, s)), r;
    }
    return (a._value = t), a;
  }
  function Mf(e, t) {
    var r, n;
    function a() {
      var s = t.apply(this, arguments);
      return s !== n && (r = (n = s) && kf(e, s)), r;
    }
    return (a._value = t), a;
  }
  function fi(e, t) {
    var r = "attr." + e;
    if (arguments.length < 2) return (r = this.tween(r)) && r._value;
    if (t == null) return this.tween(r, null);
    if (typeof t != "function") throw new Error();
    var n = Ae(e);
    return this.tween(r, (n.local ? Ef : Mf)(n, t));
  }
  function Pf(e, t) {
    return function () {
      Vt(this, e).delay = +t.apply(this, arguments);
    };
  }
  function If(e, t) {
    return (
      (t = +t),
      function () {
        Vt(this, e).delay = t;
      }
    );
  }
  function di(e) {
    var t = this._id;
    return arguments.length
      ? this.each((typeof e == "function" ? Pf : If)(t, e))
      : ne(this.node(), t).delay;
  }
  function Af(e, t) {
    return function () {
      se(this, e).duration = +t.apply(this, arguments);
    };
  }
  function Df(e, t) {
    return (
      (t = +t),
      function () {
        se(this, e).duration = t;
      }
    );
  }
  function mi(e) {
    var t = this._id;
    return arguments.length
      ? this.each((typeof e == "function" ? Af : Df)(t, e))
      : ne(this.node(), t).duration;
  }
  function Lf(e, t) {
    if (typeof t != "function") throw new Error();
    return function () {
      se(this, e).ease = t;
    };
  }
  function pi(e) {
    var t = this._id;
    return arguments.length ? this.each(Lf(t, e)) : ne(this.node(), t).ease;
  }
  function Of(e, t) {
    return function () {
      var r = t.apply(this, arguments);
      if (typeof r != "function") throw new Error();
      se(this, e).ease = r;
    };
  }
  function hi(e) {
    if (typeof e != "function") throw new Error();
    return this.each(Of(this._id, e));
  }
  function gi(e) {
    typeof e != "function" && (e = Pt(e));
    for (
      var t = this._groups, r = t.length, n = new Array(r), a = 0;
      a < r;
      ++a
    )
      for (var s = t[a], i = s.length, l = (n[a] = []), u, f = 0; f < i; ++f)
        (u = s[f]) && e.call(u, u.__data__, f, s) && l.push(u);
    return new ue(n, this._parents, this._name, this._id);
  }
  function vi(e) {
    if (e._id !== this._id) throw new Error();
    for (
      var t = this._groups,
        r = e._groups,
        n = t.length,
        a = r.length,
        s = Math.min(n, a),
        i = new Array(n),
        l = 0;
      l < s;
      ++l
    )
      for (
        var u = t[l],
          f = r[l],
          m = u.length,
          c = (i[l] = new Array(m)),
          d,
          p = 0;
        p < m;
        ++p
      )
        (d = u[p] || f[p]) && (c[p] = d);
    for (; l < n; ++l) i[l] = t[l];
    return new ue(i, this._parents, this._name, this._id);
  }
  function Hf(e) {
    return (e + "")
      .trim()
      .split(/^|\s+/)
      .every(function (t) {
        var r = t.indexOf(".");
        return r >= 0 && (t = t.slice(0, r)), !t || t === "start";
      });
  }
  function Ff(e, t, r) {
    var n,
      a,
      s = Hf(t) ? Vt : se;
    return function () {
      var i = s(this, e),
        l = i.on;
      l !== n && (a = (n = l).copy()).on(t, r), (i.on = a);
    };
  }
  function xi(e, t) {
    var r = this._id;
    return arguments.length < 2
      ? ne(this.node(), r).on.on(e)
      : this.each(Ff(r, e, t));
  }
  function Bf(e) {
    return function () {
      var t = this.parentNode;
      for (var r in this.__transition) if (+r !== e) return;
      t && t.removeChild(this);
    };
  }
  function bi() {
    return this.on("end.remove", Bf(this._id));
  }
  function yi(e) {
    var t = this._name,
      r = this._id;
    typeof e != "function" && (e = Je(e));
    for (
      var n = this._groups, a = n.length, s = new Array(a), i = 0;
      i < a;
      ++i
    )
      for (
        var l = n[i], u = l.length, f = (s[i] = new Array(u)), m, c, d = 0;
        d < u;
        ++d
      )
        (m = l[d]) &&
          (c = e.call(m, m.__data__, d, l)) &&
          ("__data__" in m && (c.__data__ = m.__data__),
          (f[d] = c),
          Ve(f[d], t, r, d, f, ne(m, r)));
    return new ue(s, this._parents, t, r);
  }
  function _i(e) {
    var t = this._name,
      r = this._id;
    typeof e != "function" && (e = Mt(e));
    for (var n = this._groups, a = n.length, s = [], i = [], l = 0; l < a; ++l)
      for (var u = n[l], f = u.length, m, c = 0; c < f; ++c)
        if ((m = u[c])) {
          for (
            var d = e.call(m, m.__data__, c, u),
              p,
              x = ne(m, r),
              b = 0,
              v = d.length;
            b < v;
            ++b
          )
            (p = d[b]) && Ve(p, t, r, b, d, x);
          s.push(d), i.push(m);
        }
    return new ue(s, i, t, r);
  }
  var zf = De.prototype.constructor;
  function wi() {
    return new zf(this._groups, this._parents);
  }
  function qf(e, t) {
    var r, n, a;
    return function () {
      var s = qe(this, e),
        i = (this.style.removeProperty(e), qe(this, e));
      return s === i
        ? null
        : s === r && i === n
        ? a
        : (a = t((r = s), (n = i)));
    };
  }
  function Ni(e) {
    return function () {
      this.style.removeProperty(e);
    };
  }
  function $f(e, t, r) {
    var n,
      a = r + "",
      s;
    return function () {
      var i = qe(this, e);
      return i === a ? null : i === n ? s : (s = t((n = i), r));
    };
  }
  function Vf(e, t, r) {
    var n, a, s;
    return function () {
      var i = qe(this, e),
        l = r(this),
        u = l + "";
      return (
        l == null && (u = l = (this.style.removeProperty(e), qe(this, e))),
        i === u ? null : i === n && u === a ? s : ((a = u), (s = t((n = i), l)))
      );
    };
  }
  function Gf(e, t) {
    var r,
      n,
      a,
      s = "style." + t,
      i = "end." + s,
      l;
    return function () {
      var u = se(this, e),
        f = u.on,
        m = u.value[s] == null ? l || (l = Ni(t)) : void 0;
      (f !== r || a !== m) && (n = (r = f).copy()).on(i, (a = m)), (u.on = n);
    };
  }
  function Ci(e, t, r) {
    var n = (e += "") == "transform" ? $n : Br;
    return t == null
      ? this.styleTween(e, qf(e, n)).on("end.style." + e, Ni(e))
      : typeof t == "function"
      ? this.styleTween(e, Vf(e, n, ht(this, "style." + e, t))).each(
          Gf(this._id, e)
        )
      : this.styleTween(e, $f(e, n, t), r).on("end.style." + e, null);
  }
  function Uf(e, t, r) {
    return function (n) {
      this.style.setProperty(e, t.call(this, n), r);
    };
  }
  function jf(e, t, r) {
    var n, a;
    function s() {
      var i = t.apply(this, arguments);
      return i !== a && (n = (a = i) && Uf(e, i, r)), n;
    }
    return (s._value = t), s;
  }
  function Ti(e, t, r) {
    var n = "style." + (e += "");
    if (arguments.length < 2) return (n = this.tween(n)) && n._value;
    if (t == null) return this.tween(n, null);
    if (typeof t != "function") throw new Error();
    return this.tween(n, jf(e, t, r ?? ""));
  }
  function Wf(e) {
    return function () {
      this.textContent = e;
    };
  }
  function Yf(e) {
    return function () {
      var t = e(this);
      this.textContent = t ?? "";
    };
  }
  function Ri(e) {
    return this.tween(
      "text",
      typeof e == "function"
        ? Yf(ht(this, "text", e))
        : Wf(e == null ? "" : e + "")
    );
  }
  function Xf(e) {
    return function (t) {
      this.textContent = e.call(this, t);
    };
  }
  function Qf(e) {
    var t, r;
    function n() {
      var a = e.apply(this, arguments);
      return a !== r && (t = (r = a) && Xf(a)), t;
    }
    return (n._value = e), n;
  }
  function ki(e) {
    var t = "text";
    if (arguments.length < 1) return (t = this.tween(t)) && t._value;
    if (e == null) return this.tween(t, null);
    if (typeof e != "function") throw new Error();
    return this.tween(t, Qf(e));
  }
  function Si() {
    for (
      var e = this._name,
        t = this._id,
        r = zr(),
        n = this._groups,
        a = n.length,
        s = 0;
      s < a;
      ++s
    )
      for (var i = n[s], l = i.length, u, f = 0; f < l; ++f)
        if ((u = i[f])) {
          var m = ne(u, t);
          Ve(u, e, r, f, i, {
            time: m.time + m.delay + m.duration,
            delay: 0,
            duration: m.duration,
            ease: m.ease,
          });
        }
    return new ue(n, this._parents, e, r);
  }
  function Ei() {
    var e,
      t,
      r = this,
      n = r._id,
      a = r.size();
    return new Promise(function (s, i) {
      var l = { value: i },
        u = {
          value: function () {
            --a === 0 && s();
          },
        };
      r.each(function () {
        var f = se(this, n),
          m = f.on;
        m !== e &&
          ((t = (e = m).copy()),
          t._.cancel.push(l),
          t._.interrupt.push(l),
          t._.end.push(u)),
          (f.on = t);
      }),
        a === 0 && s();
    });
  }
  var Kf = 0;
  function ue(e, t, r, n) {
    (this._groups = e), (this._parents = t), (this._name = r), (this._id = n);
  }
  function Mi(e) {
    return De().transition(e);
  }
  function zr() {
    return ++Kf;
  }
  var Le = De.prototype;
  ue.prototype = Mi.prototype = {
    constructor: ue,
    select: yi,
    selectAll: _i,
    selectChild: Le.selectChild,
    selectChildren: Le.selectChildren,
    filter: gi,
    merge: vi,
    selection: wi,
    transition: Si,
    call: Le.call,
    nodes: Le.nodes,
    node: Le.node,
    size: Le.size,
    empty: Le.empty,
    each: Le.each,
    on: xi,
    attr: ui,
    attrTween: fi,
    style: Ci,
    styleTween: Ti,
    text: Ri,
    textTween: ki,
    remove: bi,
    tween: ci,
    delay: di,
    duration: mi,
    ease: pi,
    easeVarying: hi,
    end: Ei,
    [Symbol.iterator]: Le[Symbol.iterator],
  };
  function qr(e) {
    return ((e *= 2) <= 1 ? e * e * e : (e -= 2) * e * e + 2) / 2;
  }
  var Jf = { time: null, delay: 0, duration: 250, ease: qr };
  function Zf(e, t) {
    for (var r; !(r = e.__transition) || !(r = r[t]); )
      if (!(e = e.parentNode)) throw new Error(`transition ${t} not found`);
    return r;
  }
  function Pi(e) {
    var t, r;
    e instanceof ue
      ? ((t = e._id), (e = e._name))
      : ((t = zr()), ((r = Jf).time = qt()), (e = e == null ? null : e + ""));
    for (var n = this._groups, a = n.length, s = 0; s < a; ++s)
      for (var i = n[s], l = i.length, u, f = 0; f < l; ++f)
        (u = i[f]) && Ve(u, e, t, f, i, r || Zf(u, t));
    return new ue(n, this._parents, e, t);
  }
  De.prototype.interrupt = li;
  De.prototype.transition = Pi;
  var { abs: Vy, max: Gy, min: Uy } = Math;
  function Ii(e) {
    return [+e[0], +e[1]];
  }
  function ed(e) {
    return [Ii(e[0]), Ii(e[1])];
  }
  var jy = {
      name: "x",
      handles: ["w", "e"].map(Un),
      input: function (e, t) {
        return e == null
          ? null
          : [
              [+e[0], t[0][1]],
              [+e[1], t[1][1]],
            ];
      },
      output: function (e) {
        return e && [e[0][0], e[1][0]];
      },
    },
    Wy = {
      name: "y",
      handles: ["n", "s"].map(Un),
      input: function (e, t) {
        return e == null
          ? null
          : [
              [t[0][0], +e[0]],
              [t[1][0], +e[1]],
            ];
      },
      output: function (e) {
        return e && [e[0][1], e[1][1]];
      },
    },
    Yy = {
      name: "xy",
      handles: ["n", "w", "e", "s", "nw", "ne", "sw", "se"].map(Un),
      input: function (e) {
        return e == null ? null : ed(e);
      },
      output: function (e) {
        return e;
      },
    };
  function Un(e) {
    return { type: e };
  }
  function td(e) {
    var t = 0,
      r = e.children,
      n = r && r.length;
    if (!n) t = 1;
    else for (; --n >= 0; ) t += r[n].value;
    e.value = t;
  }
  function Ai() {
    return this.eachAfter(td);
  }
  function Di(e, t) {
    let r = -1;
    for (let n of this) e.call(t, n, ++r, this);
    return this;
  }
  function Li(e, t) {
    for (var r = this, n = [r], a, s, i = -1; (r = n.pop()); )
      if ((e.call(t, r, ++i, this), (a = r.children)))
        for (s = a.length - 1; s >= 0; --s) n.push(a[s]);
    return this;
  }
  function Oi(e, t) {
    for (var r = this, n = [r], a = [], s, i, l, u = -1; (r = n.pop()); )
      if ((a.push(r), (s = r.children)))
        for (i = 0, l = s.length; i < l; ++i) n.push(s[i]);
    for (; (r = a.pop()); ) e.call(t, r, ++u, this);
    return this;
  }
  function Hi(e, t) {
    let r = -1;
    for (let n of this) if (e.call(t, n, ++r, this)) return n;
  }
  function Fi(e) {
    return this.eachAfter(function (t) {
      for (
        var r = +e(t.data) || 0, n = t.children, a = n && n.length;
        --a >= 0;

      )
        r += n[a].value;
      t.value = r;
    });
  }
  function Bi(e) {
    return this.eachBefore(function (t) {
      t.children && t.children.sort(e);
    });
  }
  function zi(e) {
    for (var t = this, r = rd(t, e), n = [t]; t !== r; )
      (t = t.parent), n.push(t);
    for (var a = n.length; e !== r; ) n.splice(a, 0, e), (e = e.parent);
    return n;
  }
  function rd(e, t) {
    if (e === t) return e;
    var r = e.ancestors(),
      n = t.ancestors(),
      a = null;
    for (e = r.pop(), t = n.pop(); e === t; )
      (a = e), (e = r.pop()), (t = n.pop());
    return a;
  }
  function qi() {
    for (var e = this, t = [e]; (e = e.parent); ) t.push(e);
    return t;
  }
  function $i() {
    return Array.from(this);
  }
  function Vi() {
    var e = [];
    return (
      this.eachBefore(function (t) {
        t.children || e.push(t);
      }),
      e
    );
  }
  function Gi() {
    var e = this,
      t = [];
    return (
      e.each(function (r) {
        r !== e && t.push({ source: r.parent, target: r });
      }),
      t
    );
  }
  function* Ui() {
    var e = this,
      t,
      r = [e],
      n,
      a,
      s;
    do
      for (t = r.reverse(), r = []; (e = t.pop()); )
        if ((yield e, (n = e.children)))
          for (a = 0, s = n.length; a < s; ++a) r.push(n[a]);
    while (r.length);
  }
  function gt(e, t) {
    e instanceof Map
      ? ((e = [void 0, e]), t === void 0 && (t = ad))
      : t === void 0 && (t = od);
    for (var r = new Gt(e), n, a = [r], s, i, l, u; (n = a.pop()); )
      if ((i = t(n.data)) && (u = (i = Array.from(i)).length))
        for (n.children = i, l = u - 1; l >= 0; --l)
          a.push((s = i[l] = new Gt(i[l]))),
            (s.parent = n),
            (s.depth = n.depth + 1);
    return r.eachBefore(id);
  }
  function nd() {
    return gt(this).eachBefore(sd);
  }
  function od(e) {
    return e.children;
  }
  function ad(e) {
    return Array.isArray(e) ? e[1] : null;
  }
  function sd(e) {
    e.data.value !== void 0 && (e.value = e.data.value), (e.data = e.data.data);
  }
  function id(e) {
    var t = 0;
    do e.height = t;
    while ((e = e.parent) && e.height < ++t);
  }
  function Gt(e) {
    (this.data = e), (this.depth = this.height = 0), (this.parent = null);
  }
  Gt.prototype = gt.prototype = {
    constructor: Gt,
    count: Ai,
    each: Di,
    eachAfter: Oi,
    eachBefore: Li,
    find: Hi,
    sum: Fi,
    sort: Bi,
    path: zi,
    ancestors: qi,
    descendants: $i,
    leaves: Vi,
    links: Gi,
    copy: nd,
    [Symbol.iterator]: Ui,
  };
  function ji(e) {
    (e.x0 = Math.round(e.x0)),
      (e.y0 = Math.round(e.y0)),
      (e.x1 = Math.round(e.x1)),
      (e.y1 = Math.round(e.y1));
  }
  function Wi(e, t, r, n, a) {
    for (
      var s = e.children,
        i,
        l = -1,
        u = s.length,
        f = e.value && (n - t) / e.value;
      ++l < u;

    )
      (i = s[l]), (i.y0 = r), (i.y1 = a), (i.x0 = t), (i.x1 = t += i.value * f);
  }
  function jn() {
    var e = 1,
      t = 1,
      r = 0,
      n = !1;
    function a(i) {
      var l = i.height + 1;
      return (
        (i.x0 = i.y0 = r),
        (i.x1 = e),
        (i.y1 = t / l),
        i.eachBefore(s(t, l)),
        n && i.eachBefore(ji),
        i
      );
    }
    function s(i, l) {
      return function (u) {
        u.children &&
          Wi(u, u.x0, (i * (u.depth + 1)) / l, u.x1, (i * (u.depth + 2)) / l);
        var f = u.x0,
          m = u.y0,
          c = u.x1 - r,
          d = u.y1 - r;
        c < f && (f = c = (f + c) / 2),
          d < m && (m = d = (m + d) / 2),
          (u.x0 = f),
          (u.y0 = m),
          (u.x1 = c),
          (u.y1 = d);
      };
    }
    return (
      (a.round = function (i) {
        return arguments.length ? ((n = !!i), a) : n;
      }),
      (a.size = function (i) {
        return arguments.length ? ((e = +i[0]), (t = +i[1]), a) : [e, t];
      }),
      (a.padding = function (i) {
        return arguments.length ? ((r = +i), a) : r;
      }),
      a
    );
  }
  function Yi(e, t) {
    switch (arguments.length) {
      case 0:
        break;
      case 1:
        this.range(e);
        break;
      default:
        this.range(t).domain(e);
        break;
    }
    return this;
  }
  var Wn = Symbol("implicit");
  function Ut() {
    var e = new ft(),
      t = [],
      r = [],
      n = Wn;
    function a(s) {
      let i = e.get(s);
      if (i === void 0) {
        if (n !== Wn) return n;
        e.set(s, (i = t.push(s) - 1));
      }
      return r[i % r.length];
    }
    return (
      (a.domain = function (s) {
        if (!arguments.length) return t.slice();
        (t = []), (e = new ft());
        for (let i of s) e.has(i) || e.set(i, t.push(i) - 1);
        return a;
      }),
      (a.range = function (s) {
        return arguments.length ? ((r = Array.from(s)), a) : r.slice();
      }),
      (a.unknown = function (s) {
        return arguments.length ? ((n = s), a) : n;
      }),
      (a.copy = function () {
        return Ut(t, r).unknown(n);
      }),
      Yi.apply(a, arguments),
      a
    );
  }
  function Xi(e) {
    for (var t = (e.length / 6) | 0, r = new Array(t), n = 0; n < t; )
      r[n] = "#" + e.slice(n * 6, ++n * 6);
    return r;
  }
  var Yn = Xi("4e79a7f28e2ce1575976b7b259a14fedc949af7aa1ff9da79c755fbab0ab");
  function Ge(e, t, r) {
    (this.k = e), (this.x = t), (this.y = r);
  }
  Ge.prototype = {
    constructor: Ge,
    scale: function (e) {
      return e === 1 ? this : new Ge(this.k * e, this.x, this.y);
    },
    translate: function (e, t) {
      return (e === 0) & (t === 0)
        ? this
        : new Ge(this.k, this.x + this.k * e, this.y + this.k * t);
    },
    apply: function (e) {
      return [e[0] * this.k + this.x, e[1] * this.k + this.y];
    },
    applyX: function (e) {
      return e * this.k + this.x;
    },
    applyY: function (e) {
      return e * this.k + this.y;
    },
    invert: function (e) {
      return [(e[0] - this.x) / this.k, (e[1] - this.y) / this.k];
    },
    invertX: function (e) {
      return (e - this.x) / this.k;
    },
    invertY: function (e) {
      return (e - this.y) / this.k;
    },
    rescaleX: function (e) {
      return e
        .copy()
        .domain(e.range().map(this.invertX, this).map(e.invert, e));
    },
    rescaleY: function (e) {
      return e
        .copy()
        .domain(e.range().map(this.invertY, this).map(e.invert, e));
    },
    toString: function () {
      return "translate(" + this.x + "," + this.y + ") scale(" + this.k + ")";
    },
  };
  var Xn = new Ge(1, 0, 0);
  Qn.prototype = Ge.prototype;
  function Qn(e) {
    for (; !e.__zoom; ) if (!(e = e.parentNode)) return Xn;
    return e.__zoom;
  }
  function Qi({ data: e, width: t = 900, height: r = 400 }) {
    let n = $(null),
      a = $(null);
    return (
      q(() => {
        if (!e || !n.current) return;
        let s = An(n.current);
        s.selectAll("*").remove();
        let i = 20,
          l = r || 400,
          u = gt(e)
            .sum((d) => d.value || 0)
            .sort((d, p) => (p.value || 0) - (d.value || 0));
        jn().size([t, l]).padding(1).round(!0)(u);
        let m = Ut(Yn),
          c = s
            .selectAll("g")
            .data(u.descendants())
            .join("g")
            .attr("transform", (d) => `translate(${d.x0},${d.depth * i})`);
        c
          .append("rect")
          .attr("x", 0)
          .attr("width", (d) => Math.max(0, d.x1 - d.x0))
          .attr("height", i - 1)
          .attr("fill", (d) => (d.depth ? m(d.data.name) : "#f3f4f6"))
          .style("stroke", "#fff")
          .style("cursor", "pointer")
          .on("mouseover", function (d, p) {
            if (a.current) {
              let x = (((p.value || 0) / (u.value || 1)) * 100).toFixed(2);
              (a.current.innerHTML = `
            <div style="font-weight: bold;">${p.data.name}</div>
            <div>${x}% of total</div>
            <div>Value: ${p.value}</div>
          `),
                (a.current.style.display = "block"),
                (a.current.style.left = d.pageX + 10 + "px"),
                (a.current.style.top = d.pageY - 28 + "px");
            }
          })
          .on("mousemove", function (d) {
            a.current &&
              ((a.current.style.left = d.pageX + 10 + "px"),
              (a.current.style.top = d.pageY - 28 + "px"));
          })
          .on("mouseout", function () {
            a.current && (a.current.style.display = "none");
          }),
          c
            .append("text")
            .attr("x", 4)
            .attr("y", i / 2)
            .attr("dy", "0.32em")
            .text((d) => {
              let p = d.x1 - d.x0;
              if (p < 30) return "";
              let x = d.data.name,
                b = Math.floor(p / 7);
              return x.length > b ? x.substring(0, b - 1) + "\u2026" : x;
            })
            .style("pointer-events", "none")
            .style("fill", (d) => (d.depth ? "#fff" : "#000"))
            .style("font-size", "12px")
            .style("font-family", "monospace");
      }, [e, t, r]),
      e
        ? o("div", {
            className: "relative",
            children: [
              o("svg", {
                ref: n,
                width: t,
                height: r,
                style: { width: "100%", height: "auto" },
                viewBox: `0 0 ${t} ${r}`,
              }),
              o("div", {
                ref: a,
                className:
                  "absolute bg-gray-900 text-white p-2 rounded shadow-lg text-sm",
                style: {
                  display: "none",
                  pointerEvents: "none",
                  zIndex: 1e3,
                  position: "fixed",
                },
              }),
            ],
          })
        : o("div", {
            className:
              "flex items-center justify-center h-64 text-muted-foreground",
            children: "No flame graph data available",
          })
    );
  }
  function Ki({ request: e, open: t, onOpenChange: r }) {
    let [n, a] = P(!1),
      [s, i] = P(null),
      [l, u] = P(null),
      [f, m] = P(!1),
      [c, d] = P("overview");
    q(() => {
      t && e?.ID && (p(e.ID), d("overview"));
    }, [t, e?.ID]);
    let p = async (_) => {
        m(!0);
        try {
          let T = await fe.getMetrics(_);
          i(T);
        } catch (T) {
          console.error("Failed to load metrics:", T), i(null);
        } finally {
          m(!1);
        }
      },
      x = async () => {
        if (e?.ID)
          try {
            let _ = await fe.getFlameGraph(e.ID);
            u(_);
          } catch (_) {
            console.error("Failed to load flame graph:", _);
          }
      },
      b = (_) =>
        _
          ? Object.entries(_).map(([T, C]) => `${T}: ${C.join(", ")}`).join(`
`)
          : "No headers",
      v = (_) => {
        if (!_) return "No body";
        try {
          let T = JSON.parse(_);
          return JSON.stringify(T, null, 2);
        } catch {
          return _;
        }
      },
      y = (_) =>
        _ < 1 ? "<1ms" : _ < 1e3 ? `${_}ms` : `${(_ / 1e3).toFixed(2)}s`,
      N = (_) => {
        if (!_) return "0ms";
        let T = _ / 1e6;
        return T < 1
          ? Math.round(_ / 1e3) + "\u03BCs"
          : T < 1e3
          ? T.toFixed(2) + "ms"
          : (T / 1e3).toFixed(2) + "s";
      },
      R = (_) => {
        if (!_) return "0 B";
        let T = ["B", "KB", "MB", "GB"],
          C = Math.floor(Math.log(_) / Math.log(1024));
        return Math.round((_ / Math.pow(1024, C)) * 100) / 100 + " " + T[C];
      },
      M = (_) =>
        _ >= 200 && _ < 300
          ? "default"
          : _ >= 300 && _ < 400
          ? "secondary"
          : "outline";
    if (!e) return null;
    let w = !!e.PerformanceMetrics || !!s;
    return o(Ta, {
      open: t,
      onOpenChange: r,
      children: o(Ra, {
        onClose: () => r(!1),
        isFullscreen: n,
        onToggleFullscreen: () => a(!n),
        children: [
          o("div", {
            className:
              "bg-muted/30 rounded-lg p-4 mb-6 animate-in fade-in-50 duration-300",
            children: o("div", {
              className: "flex items-center justify-between",
              children: [
                o("div", {
                  className: "flex items-center gap-4",
                  children: [
                    o(ie, {
                      className: "px-3 py-1",
                      variant: "secondary",
                      children: e.Method,
                    }),
                    o("span", {
                      className: "font-mono text-sm",
                      children: e.Path,
                    }),
                    o(ie, { variant: M(e.StatusCode), children: e.StatusCode }),
                  ],
                }),
                o("div", {
                  className:
                    "flex items-center gap-6 text-sm text-muted-foreground",
                  children: [
                    o("span", {
                      className: "flex items-center gap-1",
                      children: [
                        o("svg", {
                          className: "w-4 h-4",
                          fill: "none",
                          stroke: "currentColor",
                          viewBox: "0 0 24 24",
                          children: o("path", {
                            strokeLinecap: "round",
                            strokeLinejoin: "round",
                            strokeWidth: 2,
                            d: "M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z",
                          }),
                        }),
                        y(e.Duration),
                      ],
                    }),
                    o("span", {
                      children: new Date(e.Timestamp).toLocaleString(),
                    }),
                  ],
                }),
              ],
            }),
          }),
          o(ke, {
            value: c,
            onValueChange: d,
            className: "w-full",
            children: [
              o(_e, {
                className: "grid w-full grid-cols-5 mb-6",
                children: [
                  o(Y, {
                    value: "overview",
                    className:
                      "data-[state=active]:bg-primary data-[state=active]:text-primary-foreground",
                    children: "Overview",
                  }),
                  o(Y, { value: "headers", children: "Headers" }),
                  o(Y, { value: "body", children: "Body" }),
                  o(Y, { value: "trace", children: "Trace" }),
                  w && o(Y, { value: "performance", children: "Performance" }),
                ],
              }),
              o(X, {
                value: "overview",
                className: "space-y-6 animate-in fade-in-50 duration-300",
                children: [
                  o("div", {
                    className: "grid grid-cols-2 md:grid-cols-4 gap-4",
                    children: [
                      o(I, {
                        children: o(A, {
                          className: "p-4",
                          children: [
                            o("p", {
                              className: "text-xs text-muted-foreground mb-1",
                              children: "Request ID",
                            }),
                            o("p", {
                              className: "font-mono text-xs truncate",
                              children: e.ID,
                            }),
                          ],
                        }),
                      }),
                      o(I, {
                        children: o(A, {
                          className: "p-4",
                          children: [
                            o("p", {
                              className: "text-xs text-muted-foreground mb-1",
                              children: "Duration",
                            }),
                            o("p", {
                              className: "text-lg font-semibold",
                              children: y(e.Duration),
                            }),
                          ],
                        }),
                      }),
                      o(I, {
                        children: o(A, {
                          className: "p-4",
                          children: [
                            o("p", {
                              className: "text-xs text-muted-foreground mb-1",
                              children: "Response Size",
                            }),
                            o("p", {
                              className: "text-lg font-semibold",
                              children: R(e.ResponseBody?.length || 0),
                            }),
                          ],
                        }),
                      }),
                      o(I, {
                        children: o(A, {
                          className: "p-4",
                          children: [
                            o("p", {
                              className: "text-xs text-muted-foreground mb-1",
                              children: "Query String",
                            }),
                            o("p", {
                              className: "font-mono text-xs truncate",
                              children: e.Query || "No query",
                            }),
                          ],
                        }),
                      }),
                    ],
                  }),
                  o(I, {
                    children: o(A, {
                      className: "p-6",
                      children: [
                        o("h4", {
                          className: "text-sm font-medium mb-4",
                          children: "Request Timeline",
                        }),
                        o("div", {
                          className:
                            "relative h-12 bg-muted rounded-lg overflow-hidden",
                          children: o("div", {
                            className:
                              "absolute h-full bg-gradient-to-r from-primary/20 to-primary/60",
                            style: { width: "100%" },
                            children: o("div", {
                              className: "h-full flex items-center px-3",
                              children: o("span", {
                                className: "text-xs text-foreground/70",
                                children: y(e.Duration),
                              }),
                            }),
                          }),
                        }),
                      ],
                    }),
                  }),
                  e.Error &&
                    o(I, {
                      className: "border-destructive/50 bg-destructive/5",
                      children: o(A, {
                        className: "p-4",
                        children: o("div", {
                          className: "flex items-start gap-2",
                          children: [
                            o("svg", {
                              className: "w-5 h-5 text-destructive mt-0.5",
                              fill: "none",
                              stroke: "currentColor",
                              viewBox: "0 0 24 24",
                              children: o("path", {
                                strokeLinecap: "round",
                                strokeLinejoin: "round",
                                strokeWidth: 2,
                                d: "M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z",
                              }),
                            }),
                            o("div", {
                              className: "flex-1",
                              children: [
                                o("p", {
                                  className:
                                    "font-medium text-destructive mb-1",
                                  children: "Error Occurred",
                                }),
                                o("p", {
                                  className: "text-sm text-muted-foreground",
                                  children: e.Error,
                                }),
                              ],
                            }),
                          ],
                        }),
                      }),
                    }),
                ],
              }),
              o(X, {
                value: "headers",
                className: "space-y-6 animate-in fade-in-50 duration-300",
                children: [
                  o(I, {
                    children: o(A, {
                      className: "p-6",
                      children: [
                        o("h4", {
                          className: "text-sm font-medium mb-3",
                          children: "Request Headers",
                        }),
                        o("pre", {
                          className:
                            "bg-muted/50 p-4 rounded-lg text-xs overflow-x-auto font-mono",
                          children: b(e.RequestHeaders),
                        }),
                      ],
                    }),
                  }),
                  o(I, {
                    children: o(A, {
                      className: "p-6",
                      children: [
                        o("h4", {
                          className: "text-sm font-medium mb-3",
                          children: "Response Headers",
                        }),
                        o("pre", {
                          className:
                            "bg-muted/50 p-4 rounded-lg text-xs overflow-x-auto font-mono",
                          children: b(e.ResponseHeaders),
                        }),
                      ],
                    }),
                  }),
                ],
              }),
              o(X, {
                value: "body",
                className: "space-y-6 animate-in fade-in-50 duration-300",
                children: [
                  o(I, {
                    children: o(A, {
                      className: "p-6",
                      children: [
                        o("h4", {
                          className: "text-sm font-medium mb-3",
                          children: "Request Body",
                        }),
                        o("pre", {
                          className:
                            "bg-muted/50 p-4 rounded-lg text-xs overflow-x-auto max-h-96 font-mono",
                          children: v(e.RequestBody),
                        }),
                      ],
                    }),
                  }),
                  o(I, {
                    children: o(A, {
                      className: "p-6",
                      children: [
                        o("h4", {
                          className: "text-sm font-medium mb-3",
                          children: "Response Body",
                        }),
                        o("pre", {
                          className:
                            "bg-muted/50 p-4 rounded-lg text-xs overflow-x-auto max-h-96 font-mono",
                          children: v(e.ResponseBody),
                        }),
                      ],
                    }),
                  }),
                ],
              }),
              o(X, {
                value: "trace",
                className: "space-y-6 animate-in fade-in-50 duration-300",
                children: [
                  e.MiddlewareTrace &&
                    e.MiddlewareTrace.length > 0 &&
                    o(I, {
                      children: o(A, {
                        className: "p-6",
                        children: [
                          o("h4", {
                            className: "text-sm font-medium mb-4",
                            children: "Middleware Execution",
                          }),
                          o("div", {
                            className: "space-y-2",
                            children: e.MiddlewareTrace.map((_, T) =>
                              o(
                                "div",
                                {
                                  className:
                                    "flex items-center justify-between p-3 bg-muted/30 rounded-lg hover:bg-muted/50 transition-colors",
                                  children: [
                                    o("div", {
                                      className: "flex items-center gap-3",
                                      children: [
                                        o("div", {
                                          className:
                                            "w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium",
                                          children: T + 1,
                                        }),
                                        o("span", {
                                          className: "font-mono text-sm",
                                          children:
                                            _.name || `Middleware ${T + 1}`,
                                        }),
                                        _.type &&
                                          o(ie, {
                                            className: "text-xs",
                                            variant: "secondary",
                                            children: _.type,
                                          }),
                                      ],
                                    }),
                                    o("div", {
                                      className:
                                        "flex items-center gap-4 text-sm text-muted-foreground",
                                      children: [
                                        o("span", {
                                          children: y(_.duration || 0),
                                        }),
                                        o(ie, {
                                          variant:
                                            _.status === "completed"
                                              ? "default"
                                              : "outline",
                                          children: _.status || "completed",
                                        }),
                                      ],
                                    }),
                                  ],
                                },
                                T
                              )
                            ),
                          }),
                        ],
                      }),
                    }),
                  e.PerformanceMetrics?.sql_queries &&
                    e.PerformanceMetrics.sql_queries.length > 0 &&
                    o(I, {
                      children: o(A, {
                        className: "p-6",
                        children: [
                          o("h4", {
                            className: "text-sm font-medium mb-4",
                            children: [
                              "SQL Queries (",
                              e.PerformanceMetrics.sql_queries.length,
                              ")",
                            ],
                          }),
                          o("div", {
                            className: "space-y-3",
                            children: e.PerformanceMetrics.sql_queries
                              .slice(0, 5)
                              .map((_, T) =>
                                o(
                                  "div",
                                  {
                                    className: "p-3 bg-muted/30 rounded-lg",
                                    children: [
                                      o("pre", {
                                        className:
                                          "text-xs font-mono overflow-x-auto",
                                        children: _.query,
                                      }),
                                      o("div", {
                                        className:
                                          "flex items-center gap-4 mt-2 text-xs text-muted-foreground",
                                        children: [
                                          o("span", {
                                            children: [
                                              "Duration: ",
                                              N(_.duration),
                                            ],
                                          }),
                                          o("span", {
                                            children: ["Rows: ", _.rows],
                                          }),
                                          _.error &&
                                            o("span", {
                                              className: "text-red-600",
                                              children: ["Error: ", _.error],
                                            }),
                                        ],
                                      }),
                                    ],
                                  },
                                  T
                                )
                              ),
                          }),
                        ],
                      }),
                    }),
                  e.PerformanceMetrics?.http_calls &&
                    e.PerformanceMetrics.http_calls.length > 0 &&
                    o(I, {
                      children: o(A, {
                        className: "p-6",
                        children: [
                          o("h4", {
                            className: "text-sm font-medium mb-4",
                            children: [
                              "HTTP Calls (",
                              e.PerformanceMetrics.http_calls.length,
                              ")",
                            ],
                          }),
                          o("div", {
                            className: "space-y-3",
                            children: e.PerformanceMetrics.http_calls
                              .slice(0, 5)
                              .map((_, T) =>
                                o(
                                  "div",
                                  {
                                    className:
                                      "flex items-center justify-between p-3 bg-muted/30 rounded-lg",
                                    children: [
                                      o("div", {
                                        className: "flex items-center gap-3",
                                        children: [
                                          o(ie, {
                                            variant: "outline",
                                            children: _.method,
                                          }),
                                          o("span", {
                                            className: "text-sm font-mono",
                                            children: _.url,
                                          }),
                                        ],
                                      }),
                                      o("div", {
                                        className:
                                          "flex items-center gap-4 text-sm text-muted-foreground",
                                        children: [
                                          o("span", {
                                            children: ["Status: ", _.status],
                                          }),
                                          o("span", {
                                            children: N(_.duration),
                                          }),
                                        ],
                                      }),
                                    ],
                                  },
                                  T
                                )
                              ),
                          }),
                        ],
                      }),
                    }),
                  e.RouteTrace &&
                    o(I, {
                      children: o(A, {
                        className: "p-6",
                        children: [
                          o("h4", {
                            className: "text-sm font-medium mb-4",
                            children: "Route Information",
                          }),
                          o("div", {
                            className: "space-y-3",
                            children: [
                              o("div", {
                                className: "flex justify-between py-2",
                                children: [
                                  o("span", {
                                    className: "text-sm text-muted-foreground",
                                    children: "Pattern",
                                  }),
                                  o("span", {
                                    className: "font-mono text-sm",
                                    children: e.RouteTrace.pattern || e.Path,
                                  }),
                                ],
                              }),
                              o("div", {
                                className: "flex justify-between py-2",
                                children: [
                                  o("span", {
                                    className: "text-sm text-muted-foreground",
                                    children: "Handler",
                                  }),
                                  o("span", {
                                    className: "font-mono text-sm",
                                    children:
                                      e.RouteTrace.handler || "DefaultHandler",
                                  }),
                                ],
                              }),
                              o("div", {
                                className: "flex justify-between py-2",
                                children: [
                                  o("span", {
                                    className: "text-sm text-muted-foreground",
                                    children: "Match Time",
                                  }),
                                  o("span", {
                                    className: "font-mono text-sm",
                                    children: y(e.RouteTrace.matchTime || 0),
                                  }),
                                ],
                              }),
                            ],
                          }),
                        ],
                      }),
                    }),
                ],
              }),
              w &&
                o(X, {
                  value: "performance",
                  className: "space-y-6 animate-in fade-in-50 duration-300",
                  onFocus: () => {
                    c === "performance" && !l && x();
                  },
                  children: f
                    ? o("div", {
                        className: "flex items-center justify-center h-64",
                        children: o("div", {
                          className: "text-muted-foreground",
                          children: "Loading performance metrics...",
                        }),
                      })
                    : s
                    ? o(h, {
                        children: [
                          o("div", {
                            className: "grid grid-cols-2 md:grid-cols-4 gap-4",
                            children: [
                              o(I, {
                                children: o(A, {
                                  className: "p-4",
                                  children: [
                                    o("p", {
                                      className:
                                        "text-xs text-muted-foreground mb-1",
                                      children: "CPU Time",
                                    }),
                                    o("p", {
                                      className: "text-lg font-semibold",
                                      children: N(s.cpu_time),
                                    }),
                                  ],
                                }),
                              }),
                              o(I, {
                                children: o(A, {
                                  className: "p-4",
                                  children: [
                                    o("p", {
                                      className:
                                        "text-xs text-muted-foreground mb-1",
                                      children: "Memory",
                                    }),
                                    o("p", {
                                      className: "text-lg font-semibold",
                                      children: R(s.memory_alloc),
                                    }),
                                  ],
                                }),
                              }),
                              o(I, {
                                children: o(A, {
                                  className: "p-4",
                                  children: [
                                    o("p", {
                                      className:
                                        "text-xs text-muted-foreground mb-1",
                                      children: "Goroutines",
                                    }),
                                    o("p", {
                                      className: "text-lg font-semibold",
                                      children: s.num_goroutines || 0,
                                    }),
                                  ],
                                }),
                              }),
                              o(I, {
                                children: o(A, {
                                  className: "p-4",
                                  children: [
                                    o("p", {
                                      className:
                                        "text-xs text-muted-foreground mb-1",
                                      children: "GC Pauses",
                                    }),
                                    o("p", {
                                      className: "text-lg font-semibold",
                                      children: N(s.gc_pause_total),
                                    }),
                                  ],
                                }),
                              }),
                            ],
                          }),
                          s.bottlenecks &&
                            s.bottlenecks.length > 0 &&
                            o(I, {
                              children: o(A, {
                                className: "p-6",
                                children: [
                                  o("h4", {
                                    className: "text-sm font-medium mb-4",
                                    children: "Performance Bottlenecks",
                                  }),
                                  o("div", {
                                    className: "space-y-3",
                                    children: s.bottlenecks.map((_, T) =>
                                      o(
                                        "div",
                                        {
                                          className:
                                            "p-4 bg-muted/30 rounded-lg hover:bg-muted/50 transition-all duration-200 hover:shadow-sm",
                                          children: o("div", {
                                            className:
                                              "flex justify-between items-start",
                                            children: [
                                              o("div", {
                                                className: "flex-1",
                                                children: [
                                                  o("div", {
                                                    className:
                                                      "flex items-center gap-2 mb-2",
                                                    children: [
                                                      o(ie, {
                                                        variant: "secondary",
                                                        children:
                                                          _.type.toUpperCase(),
                                                      }),
                                                      o("span", {
                                                        className:
                                                          "font-medium text-sm",
                                                        children: _.description,
                                                      }),
                                                    ],
                                                  }),
                                                  o("p", {
                                                    className:
                                                      "text-xs text-muted-foreground",
                                                    children: _.suggestion,
                                                  }),
                                                ],
                                              }),
                                              o("div", {
                                                className: "text-right ml-4",
                                                children: [
                                                  o("div", {
                                                    className:
                                                      "text-lg font-bold",
                                                    children: [
                                                      (_.impact * 100).toFixed(
                                                        1
                                                      ),
                                                      "%",
                                                    ],
                                                  }),
                                                  o("div", {
                                                    className:
                                                      "text-xs text-muted-foreground",
                                                    children: N(_.duration),
                                                  }),
                                                ],
                                              }),
                                            ],
                                          }),
                                        },
                                        T
                                      )
                                    ),
                                  }),
                                ],
                              }),
                            }),
                          l &&
                            o(I, {
                              children: o(A, {
                                className: "p-6",
                                children: [
                                  o("h4", {
                                    className: "text-sm font-medium mb-4",
                                    children: "CPU Profile Flame Graph",
                                  }),
                                  o("div", {
                                    className: "w-full overflow-x-auto",
                                    children: o(Qi, {
                                      data: l,
                                      width: 900,
                                      height: 400,
                                    }),
                                  }),
                                ],
                              }),
                            }),
                        ],
                      })
                    : null,
                }),
            ],
          }),
        ],
      }),
    });
  }
  function Ji() {
    let [e, t] = P({
      goVersion: "Loading...",
      goos: "Loading...",
      goarch: "Loading...",
      hostname: "Loading...",
      cpuCores: 0,
      memoryUsed: 0,
      memoryTotal: 0,
      envVars: {},
    });
    q(() => {
      r();
    }, []);
    let r = async () => {
        try {
          let a = await fetch("/__viz/api/system-info");
          if (a.ok) {
            let s = await a.json();
            t(s);
          }
        } catch (a) {
          console.error("Failed to fetch system info:", a),
            t({
              goVersion: "go1.21.0",
              goos: "darwin",
              goarch: "arm64",
              hostname: "localhost",
              cpuCores: navigator.hardwareConcurrency || 4,
              memoryUsed: 256,
              memoryTotal: 1024,
              envVars: {
                PATH: "/usr/local/bin:/usr/bin:/bin",
                HOME: "/Users/user",
                GOPATH: "/Users/user/go",
              },
            });
        }
      },
      n = (e.memoryUsed / e.memoryTotal) * 100;
    return o("div", {
      className: "space-y-6",
      children: [
        o("div", {
          className: "grid grid-cols-1 md:grid-cols-3 gap-6",
          children: [
            o(I, {
              children: [
                o(j, { children: o(W, { children: "Go Environment" }) }),
                o(A, {
                  className: "space-y-2",
                  children: [
                    o("div", {
                      className: "flex justify-between",
                      children: [
                        o("span", {
                          className: "text-sm text-muted-foreground",
                          children: "Version:",
                        }),
                        o("span", {
                          className: "text-sm font-medium",
                          children: e.goVersion,
                        }),
                      ],
                    }),
                    o("div", {
                      className: "flex justify-between",
                      children: [
                        o("span", {
                          className: "text-sm text-muted-foreground",
                          children: "GOOS:",
                        }),
                        o("span", {
                          className: "text-sm font-medium",
                          children: e.goos,
                        }),
                      ],
                    }),
                    o("div", {
                      className: "flex justify-between",
                      children: [
                        o("span", {
                          className: "text-sm text-muted-foreground",
                          children: "GOARCH:",
                        }),
                        o("span", {
                          className: "text-sm font-medium",
                          children: e.goarch,
                        }),
                      ],
                    }),
                  ],
                }),
              ],
            }),
            o(I, {
              children: [
                o(j, { children: o(W, { children: "System" }) }),
                o(A, {
                  className: "space-y-2",
                  children: [
                    o("div", {
                      className: "flex justify-between",
                      children: [
                        o("span", {
                          className: "text-sm text-muted-foreground",
                          children: "Hostname:",
                        }),
                        o("span", {
                          className: "text-sm font-medium",
                          children: e.hostname,
                        }),
                      ],
                    }),
                    o("div", {
                      className: "flex justify-between",
                      children: [
                        o("span", {
                          className: "text-sm text-muted-foreground",
                          children: "OS:",
                        }),
                        o("span", {
                          className: "text-sm font-medium",
                          children: e.goos,
                        }),
                      ],
                    }),
                    o("div", {
                      className: "flex justify-between",
                      children: [
                        o("span", {
                          className: "text-sm text-muted-foreground",
                          children: "CPU Cores:",
                        }),
                        o("span", {
                          className: "text-sm font-medium",
                          children: e.cpuCores,
                        }),
                      ],
                    }),
                  ],
                }),
              ],
            }),
            o(I, {
              children: [
                o(j, { children: o(W, { children: "Memory Usage" }) }),
                o(A, {
                  children: o("div", {
                    className: "space-y-2",
                    children: [
                      o("div", {
                        className: "w-full bg-gray-200 rounded-full h-2",
                        children: o("div", {
                          className:
                            "bg-primary h-2 rounded-full transition-all duration-300",
                          style: { width: `${n}%` },
                        }),
                      }),
                      o("div", {
                        className: "flex justify-between text-sm",
                        children: [
                          o("span", {
                            className: "text-muted-foreground",
                            children: [
                              e.memoryUsed,
                              "MB / ",
                              e.memoryTotal,
                              "MB",
                            ],
                          }),
                          o("span", {
                            className: "font-medium",
                            children: [n.toFixed(1), "%"],
                          }),
                        ],
                      }),
                    ],
                  }),
                }),
              ],
            }),
          ],
        }),
        o(I, {
          children: [
            o(j, {
              children: [
                o(W, { children: "Environment Variables" }),
                o(Tt, {
                  children:
                    "System environment variables (sensitive values are redacted)",
                }),
              ],
            }),
            o(A, {
              children: o("div", {
                className: "max-h-96 overflow-auto",
                children: o(Rt, {
                  children: [
                    o(kt, {
                      children: o(Qe, {
                        children: [
                          o(ge, { children: "Name" }),
                          o(ge, { children: "Value" }),
                        ],
                      }),
                    }),
                    o(St, {
                      children: Object.entries(e.envVars).map(([a, s]) =>
                        o(
                          Qe,
                          {
                            children: [
                              o(ve, {
                                className: "font-mono text-sm",
                                children: a,
                              }),
                              o(ve, {
                                className:
                                  "font-mono text-sm text-muted-foreground",
                                children: s,
                              }),
                            ],
                          },
                          a
                        )
                      ),
                    }),
                  ],
                }),
              }),
            }),
          ],
        }),
      ],
    });
  }
  function Kn({ onFilterChange: e, onClear: t }) {
    let r = () => {
      let a = {
        method: document.getElementById("method-filter")?.value || "",
        statusCode: document.getElementById("status-filter")?.value || "",
        path: document.getElementById("path-filter")?.value || "",
        minDuration: document.getElementById("duration-filter")?.value || "",
      };
      e(a);
    };
    return o(I, {
      children: [
        o(j, { children: o(W, { children: "Filters" }) }),
        o(A, {
          children: [
            o("div", {
              className: "grid grid-cols-2 md:grid-cols-4 gap-4 mb-4",
              children: [
                o("div", {
                  children: [
                    o("label", {
                      htmlFor: "method-filter",
                      className: "block text-sm font-medium mb-1",
                      children: "HTTP Method",
                    }),
                    o("select", {
                      id: "method-filter",
                      className:
                        "w-full px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary",
                      onChange: r,
                      children: [
                        o("option", { value: "", children: "All Methods" }),
                        o("option", { value: "GET", children: "GET" }),
                        o("option", { value: "POST", children: "POST" }),
                        o("option", { value: "PUT", children: "PUT" }),
                        o("option", { value: "DELETE", children: "DELETE" }),
                        o("option", { value: "PATCH", children: "PATCH" }),
                        o("option", { value: "HEAD", children: "HEAD" }),
                        o("option", { value: "OPTIONS", children: "OPTIONS" }),
                      ],
                    }),
                  ],
                }),
                o("div", {
                  children: [
                    o("label", {
                      htmlFor: "status-filter",
                      className: "block text-sm font-medium mb-1",
                      children: "Status Code",
                    }),
                    o("select", {
                      id: "status-filter",
                      className:
                        "w-full px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary",
                      onChange: r,
                      children: [
                        o("option", {
                          value: "",
                          children: "All Status Codes",
                        }),
                        o("option", { value: "2xx", children: "2xx Success" }),
                        o("option", { value: "3xx", children: "3xx Redirect" }),
                        o("option", {
                          value: "4xx",
                          children: "4xx Client Error",
                        }),
                        o("option", {
                          value: "5xx",
                          children: "5xx Server Error",
                        }),
                      ],
                    }),
                  ],
                }),
                o("div", {
                  children: [
                    o("label", {
                      htmlFor: "path-filter",
                      className: "block text-sm font-medium mb-1",
                      children: "Path Contains",
                    }),
                    o("input", {
                      type: "text",
                      id: "path-filter",
                      placeholder: "Filter by path...",
                      className:
                        "w-full px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary",
                      onInput: r,
                    }),
                  ],
                }),
                o("div", {
                  children: [
                    o("label", {
                      htmlFor: "duration-filter",
                      className: "block text-sm font-medium mb-1",
                      children: "Min Duration (ms)",
                    }),
                    o("input", {
                      type: "number",
                      id: "duration-filter",
                      placeholder: "Min duration...",
                      className:
                        "w-full px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary",
                      onInput: r,
                    }),
                  ],
                }),
              ],
            }),
            o("div", {
              className: "flex gap-2",
              children: [
                o(K, { onClick: r, children: "Apply Filters" }),
                o(K, {
                  variant: "secondary",
                  onClick: () => {
                    (document.getElementById("method-filter").value = ""),
                      (document.getElementById("status-filter").value = ""),
                      (document.getElementById("path-filter").value = ""),
                      (document.getElementById("duration-filter").value = ""),
                      e({
                        method: "",
                        statusCode: "",
                        path: "",
                        minDuration: "",
                      });
                  },
                  children: "Reset",
                }),
                o(K, {
                  variant: "destructive",
                  onClick: t,
                  children: "Clear All Requests",
                }),
              ],
            }),
          ],
        }),
      ],
    });
  }
  function Zi({ requestIds: e, allRequests: t, onClose: r }) {
    let [n, a] = P([]),
      [s, i] = P(!0);
    q(() => {
      l();
    }, [e]);
    let l = async () => {
      try {
        i(!0);
        let d = await fe.compareRequests(e);
        a(d);
      } catch (d) {
        console.error("Failed to load comparison data:", d);
        let p = t.filter((x) => e.includes(x.ID));
        a(p);
      } finally {
        i(!1);
      }
    };
    if (s)
      return o("div", {
        className: "flex items-center justify-center h-64",
        children: o("div", {
          className: "text-muted-foreground",
          children: "Loading comparison...",
        }),
      });
    if (n.length === 0)
      return o("div", {
        className: "flex flex-col items-center justify-center h-64",
        children: [
          o("div", {
            className: "text-muted-foreground mb-4",
            children: "No requests found for comparison",
          }),
          o(K, { onClick: r, children: "Close" }),
        ],
      });
    let u = (d) => (d < 1e3 ? `${d}ms` : `${(d / 1e3).toFixed(2)}s`),
      f = (d) => new Date(d).toLocaleString(),
      m = (d) =>
        d >= 200 && d < 300
          ? "text-green-600"
          : d >= 300 && d < 400
          ? "text-blue-600"
          : d >= 400 && d < 500
          ? "text-yellow-600"
          : d >= 500
          ? "text-red-600"
          : "text-gray-600",
      c = (d, p) => {
        let x = p.every((b) => JSON.stringify(b) === JSON.stringify(p[0]));
        return o("tr", {
          children: [
            o("td", {
              className: "font-medium text-sm p-2 border-b",
              children: d,
            }),
            p.map((b, v) =>
              o(
                "td",
                {
                  className: D(
                    "text-sm p-2 border-b",
                    !x && "bg-yellow-50 dark:bg-yellow-900/10"
                  ),
                  children:
                    typeof b == "object" ? JSON.stringify(b, null, 2) : b,
                },
                v
              )
            ),
          ],
        });
      };
    return o("div", {
      className: "space-y-4",
      children: [
        o("div", {
          className: "flex items-center justify-between mb-4",
          children: [
            o("h2", {
              className: "text-2xl font-bold",
              children: "Request Comparison",
            }),
            o(K, { onClick: r, variant: "outline", children: "Close" }),
          ],
        }),
        o(ke, {
          defaultValue: "overview",
          className: "w-full",
          children: [
            o(_e, {
              className: "grid w-full grid-cols-4",
              children: [
                o(Y, { value: "overview", children: "Overview" }),
                o(Y, { value: "headers", children: "Headers" }),
                o(Y, { value: "body", children: "Body" }),
                o(Y, { value: "performance", children: "Performance" }),
              ],
            }),
            o(X, {
              value: "overview",
              className: "space-y-4",
              children: o(I, {
                children: [
                  o(j, { children: o(W, { children: "Request Details" }) }),
                  o(A, {
                    children: o("div", {
                      className: "overflow-x-auto",
                      children: o("table", {
                        className: "w-full",
                        children: [
                          o("thead", {
                            children: o("tr", {
                              children: [
                                o("th", {
                                  className: "text-left p-2 border-b",
                                  children: "Property",
                                }),
                                n.map((d, p) =>
                                  o(
                                    "th",
                                    {
                                      className: "text-left p-2 border-b",
                                      children: ["Request ", p + 1],
                                    },
                                    p
                                  )
                                ),
                              ],
                            }),
                          }),
                          o("tbody", {
                            children: [
                              c(
                                "Method",
                                n.map((d) => d.Method)
                              ),
                              c(
                                "Path",
                                n.map((d) => d.Path)
                              ),
                              c(
                                "Query",
                                n.map((d) => d.Query || "None")
                              ),
                              c(
                                "Status",
                                n.map((d) =>
                                  o("span", {
                                    className: m(d.StatusCode),
                                    children: d.StatusCode,
                                  })
                                )
                              ),
                              c(
                                "Duration",
                                n.map((d) => u(d.Duration))
                              ),
                              c(
                                "Timestamp",
                                n.map((d) => f(d.Timestamp))
                              ),
                            ],
                          }),
                        ],
                      }),
                    }),
                  }),
                ],
              }),
            }),
            o(X, {
              value: "headers",
              className: "space-y-4",
              children: o("div", {
                className: "grid grid-cols-1 lg:grid-cols-2 gap-4",
                children: [
                  o(I, {
                    children: [
                      o(j, { children: o(W, { children: "Request Headers" }) }),
                      o(A, {
                        children: o("div", {
                          className: "space-y-4",
                          children: n.map((d, p) =>
                            o(
                              "div",
                              {
                                children: [
                                  o("h4", {
                                    className: "font-medium mb-2",
                                    children: ["Request ", p + 1],
                                  }),
                                  o("div", {
                                    className:
                                      "bg-gray-50 dark:bg-gray-900 rounded p-2 text-xs font-mono",
                                    children: Object.entries(
                                      d.RequestHeaders || {}
                                    ).map(([x, b]) =>
                                      o(
                                        "div",
                                        {
                                          children: [
                                            o("span", {
                                              className: "text-blue-600",
                                              children: [x, ":"],
                                            }),
                                            " ",
                                            Array.isArray(b) ? b.join(", ") : b,
                                          ],
                                        },
                                        x
                                      )
                                    ),
                                  }),
                                ],
                              },
                              p
                            )
                          ),
                        }),
                      }),
                    ],
                  }),
                  o(I, {
                    children: [
                      o(j, {
                        children: o(W, { children: "Response Headers" }),
                      }),
                      o(A, {
                        children: o("div", {
                          className: "space-y-4",
                          children: n.map((d, p) =>
                            o(
                              "div",
                              {
                                children: [
                                  o("h4", {
                                    className: "font-medium mb-2",
                                    children: ["Request ", p + 1],
                                  }),
                                  o("div", {
                                    className:
                                      "bg-gray-50 dark:bg-gray-900 rounded p-2 text-xs font-mono",
                                    children: Object.entries(
                                      d.ResponseHeaders || {}
                                    ).map(([x, b]) =>
                                      o(
                                        "div",
                                        {
                                          children: [
                                            o("span", {
                                              className: "text-green-600",
                                              children: [x, ":"],
                                            }),
                                            " ",
                                            Array.isArray(b) ? b.join(", ") : b,
                                          ],
                                        },
                                        x
                                      )
                                    ),
                                  }),
                                ],
                              },
                              p
                            )
                          ),
                        }),
                      }),
                    ],
                  }),
                ],
              }),
            }),
            o(X, {
              value: "body",
              className: "space-y-4",
              children: o("div", {
                className: "grid grid-cols-1 lg:grid-cols-2 gap-4",
                children: [
                  o(I, {
                    children: [
                      o(j, { children: o(W, { children: "Request Body" }) }),
                      o(A, {
                        children: o("div", {
                          className: "space-y-4",
                          children: n.map((d, p) =>
                            o(
                              "div",
                              {
                                children: [
                                  o("h4", {
                                    className: "font-medium mb-2",
                                    children: ["Request ", p + 1],
                                  }),
                                  o("div", {
                                    className:
                                      "bg-gray-50 dark:bg-gray-900 rounded p-2",
                                    children: o("pre", {
                                      className: "text-xs overflow-x-auto",
                                      children:
                                        d.RequestBody || "No request body",
                                    }),
                                  }),
                                ],
                              },
                              p
                            )
                          ),
                        }),
                      }),
                    ],
                  }),
                  o(I, {
                    children: [
                      o(j, { children: o(W, { children: "Response Body" }) }),
                      o(A, {
                        children: o("div", {
                          className: "space-y-4",
                          children: n.map((d, p) =>
                            o(
                              "div",
                              {
                                children: [
                                  o("h4", {
                                    className: "font-medium mb-2",
                                    children: ["Request ", p + 1],
                                  }),
                                  o("div", {
                                    className:
                                      "bg-gray-50 dark:bg-gray-900 rounded p-2",
                                    children: o("pre", {
                                      className:
                                        "text-xs overflow-x-auto max-h-48 overflow-y-auto",
                                      children:
                                        d.ResponseBody || "No response body",
                                    }),
                                  }),
                                ],
                              },
                              p
                            )
                          ),
                        }),
                      }),
                    ],
                  }),
                ],
              }),
            }),
            o(X, {
              value: "performance",
              className: "space-y-4",
              children: o(I, {
                children: [
                  o(j, { children: o(W, { children: "Performance Metrics" }) }),
                  o(A, {
                    children: n.some((d) => d.PerformanceMetrics)
                      ? o("div", {
                          className: "overflow-x-auto",
                          children: o("table", {
                            className: "w-full",
                            children: [
                              o("thead", {
                                children: o("tr", {
                                  children: [
                                    o("th", {
                                      className: "text-left p-2 border-b",
                                      children: "Metric",
                                    }),
                                    n.map((d, p) =>
                                      o(
                                        "th",
                                        {
                                          className: "text-left p-2 border-b",
                                          children: ["Request ", p + 1],
                                        },
                                        p
                                      )
                                    ),
                                  ],
                                }),
                              }),
                              o("tbody", {
                                children: [
                                  c(
                                    "CPU Time",
                                    n.map((d) =>
                                      d.PerformanceMetrics
                                        ? `${d.PerformanceMetrics.cpu_time}ms`
                                        : "N/A"
                                    )
                                  ),
                                  c(
                                    "Memory Allocated",
                                    n.map((d) =>
                                      d.PerformanceMetrics
                                        ? `${(
                                            d.PerformanceMetrics.memory_alloc /
                                            1024 /
                                            1024
                                          ).toFixed(2)}MB`
                                        : "N/A"
                                    )
                                  ),
                                  c(
                                    "Goroutines",
                                    n.map(
                                      (d) =>
                                        d.PerformanceMetrics?.num_goroutines ||
                                        "N/A"
                                    )
                                  ),
                                  c(
                                    "GC Runs",
                                    n.map(
                                      (d) =>
                                        d.PerformanceMetrics?.num_gc || "N/A"
                                    )
                                  ),
                                  c(
                                    "GC Pause",
                                    n.map((d) =>
                                      d.PerformanceMetrics
                                        ? `${d.PerformanceMetrics.gc_pause_total}ms`
                                        : "N/A"
                                    )
                                  ),
                                ],
                              }),
                            ],
                          }),
                        })
                      : o("div", {
                          className: "text-center text-muted-foreground py-8",
                          children:
                            "No performance metrics available for these requests",
                        }),
                  }),
                ],
              }),
            }),
          ],
        }),
      ],
    });
  }
  var jt = H(({ className: e, type: t, ...r }, n) =>
    o("input", {
      type: t,
      className: D(
        "flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-base ring-offset-white file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-slate-950 placeholder:text-slate-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-950 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm dark:border-slate-800 dark:bg-slate-950 dark:ring-offset-slate-950 dark:file:text-slate-50 dark:placeholder:text-slate-400 dark:focus-visible:ring-slate-300",
        e
      ),
      ref: n,
      ...r,
    })
  );
  jt.displayName = "Input";
  function el({ request: e, onClose: t }) {
    let [r, n] = P(e.Path + (e.Query ? `?${e.Query}` : "")),
      [a, s] = P(e.Method),
      [i, l] = P(() => {
        let w = {};
        return (
          e.RequestHeaders &&
            Object.entries(e.RequestHeaders).forEach(([_, T]) => {
              w[_] = Array.isArray(T) ? T[0] : T;
            }),
          w
        );
      }),
      [u, f] = P(e.RequestBody || ""),
      [m, c] = P(!1),
      [d, p] = P(null),
      [x, b] = P(null),
      v = async () => {
        try {
          c(!0), b(null);
          let w = r;
          if (!w.startsWith("http")) {
            let T = e.RequestHeaders?.Host;
            w =
              "http://" + (T ? (Array.isArray(T) ? T[0] : T) : "localhost") + w;
          }
          let _ = await fe.replayRequest({
            requestId: e.ID,
            url: w,
            method: a,
            headers: i,
            body: u,
          });
          p(_);
        } catch (w) {
          b(w.message || "Failed to replay request");
        } finally {
          c(!1);
        }
      },
      y = (w, _) => {
        l((T) => ({ ...T, [w]: _ }));
      },
      N = () => {
        let w = prompt("Enter header name:");
        w && l((_) => ({ ..._, [w]: "" }));
      },
      R = (w) => {
        l((_) => {
          let T = { ..._ };
          return delete T[w], T;
        });
      },
      M = (w) =>
        w >= 200 && w < 300
          ? "bg-green-100 text-green-800"
          : w >= 300 && w < 400
          ? "bg-blue-100 text-blue-800"
          : w >= 400 && w < 500
          ? "bg-yellow-100 text-yellow-800"
          : w >= 500
          ? "bg-red-100 text-red-800"
          : "bg-gray-100 text-gray-800";
    return o("div", {
      className: "space-y-4",
      children: [
        o("div", {
          className: "flex items-center justify-between mb-4",
          children: [
            o("h2", {
              className: "text-2xl font-bold",
              children: "Replay Request",
            }),
            o(K, { onClick: t, variant: "outline", children: "Close" }),
          ],
        }),
        o(I, {
          children: [
            o(j, { children: o(W, { children: "Request Configuration" }) }),
            o(A, {
              className: "space-y-4",
              children: [
                o("div", {
                  className: "grid grid-cols-1 md:grid-cols-2 gap-4",
                  children: [
                    o("div", {
                      children: [
                        o("label", {
                          className: "text-sm font-medium mb-1 block",
                          children: "Method",
                        }),
                        o("select", {
                          value: a,
                          onChange: (w) => s(w.target.value),
                          className: "w-full p-2 border rounded-md",
                          children: [
                            o("option", { value: "GET", children: "GET" }),
                            o("option", { value: "POST", children: "POST" }),
                            o("option", { value: "PUT", children: "PUT" }),
                            o("option", { value: "PATCH", children: "PATCH" }),
                            o("option", {
                              value: "DELETE",
                              children: "DELETE",
                            }),
                            o("option", { value: "HEAD", children: "HEAD" }),
                            o("option", {
                              value: "OPTIONS",
                              children: "OPTIONS",
                            }),
                          ],
                        }),
                      ],
                    }),
                    o("div", {
                      children: [
                        o("label", {
                          className: "text-sm font-medium mb-1 block",
                          children: "URL",
                        }),
                        o(jt, {
                          value: r,
                          onChange: (w) => n(w.target.value),
                          placeholder: "Enter URL",
                        }),
                      ],
                    }),
                  ],
                }),
                o("div", {
                  children: [
                    o("div", {
                      className: "flex items-center justify-between mb-2",
                      children: [
                        o("label", {
                          className: "text-sm font-medium",
                          children: "Headers",
                        }),
                        o(K, {
                          size: "sm",
                          variant: "outline",
                          onClick: N,
                          children: "Add Header",
                        }),
                      ],
                    }),
                    o("div", {
                      className: "space-y-2 max-h-48 overflow-y-auto",
                      children: Object.entries(i).map(([w, _]) =>
                        o(
                          "div",
                          {
                            className: "flex items-center gap-2",
                            children: [
                              o(jt, {
                                value: w,
                                disabled: !0,
                                className: "flex-1 font-mono text-sm",
                              }),
                              o(jt, {
                                value: _,
                                onChange: (T) => y(w, T.target.value),
                                placeholder: "Value",
                                className: "flex-2 font-mono text-sm",
                              }),
                              o(K, {
                                size: "sm",
                                variant: "ghost",
                                onClick: () => R(w),
                                className: "text-red-600 hover:text-red-700",
                                children: "Remove",
                              }),
                            ],
                          },
                          w
                        )
                      ),
                    }),
                  ],
                }),
                (a === "POST" || a === "PUT" || a === "PATCH") &&
                  o("div", {
                    children: [
                      o("label", {
                        className: "text-sm font-medium mb-1 block",
                        children: "Request Body",
                      }),
                      o("textarea", {
                        value: u,
                        onChange: (w) => f(w.target.value),
                        className:
                          "w-full p-2 border rounded-md font-mono text-sm",
                        rows: 6,
                        placeholder: "Enter request body (JSON, XML, etc.)",
                      }),
                    ],
                  }),
                o("div", {
                  className: "flex justify-end gap-2",
                  children: [
                    o(K, {
                      onClick: t,
                      variant: "outline",
                      children: "Cancel",
                    }),
                    o(K, {
                      onClick: v,
                      disabled: m,
                      className: D(m && "opacity-50 cursor-not-allowed"),
                      children: m ? "Replaying..." : "Send Request",
                    }),
                  ],
                }),
              ],
            }),
          ],
        }),
        x &&
          o(I, {
            className: "border-red-200 bg-red-50",
            children: [
              o(j, {
                children: o(W, {
                  className: "text-red-800",
                  children: "Error",
                }),
              }),
              o(A, {
                children: o("p", { className: "text-red-700", children: x }),
              }),
            ],
          }),
        d &&
          o(I, {
            children: [
              o(j, { children: o(W, { children: "Response" }) }),
              o(A, {
                children: o(ke, {
                  defaultValue: "overview",
                  className: "w-full",
                  children: [
                    o(_e, {
                      children: [
                        o(Y, { value: "overview", children: "Overview" }),
                        o(Y, { value: "headers", children: "Headers" }),
                        o(Y, { value: "body", children: "Body" }),
                      ],
                    }),
                    o(X, {
                      value: "overview",
                      className: "space-y-4",
                      children: o("div", {
                        className: "grid grid-cols-1 md:grid-cols-3 gap-4",
                        children: [
                          o("div", {
                            children: [
                              o("span", {
                                className: "text-sm text-muted-foreground",
                                children: "Status",
                              }),
                              o("div", {
                                className: "mt-1",
                                children: o(ie, {
                                  className: M(d.statusCode),
                                  children: d.statusCode,
                                }),
                              }),
                            ],
                          }),
                          o("div", {
                            children: [
                              o("span", {
                                className: "text-sm text-muted-foreground",
                                children: "Duration",
                              }),
                              o("div", {
                                className: "mt-1 text-lg font-medium",
                                children: [d.duration, "ms"],
                              }),
                            ],
                          }),
                          o("div", {
                            children: [
                              o("span", {
                                className: "text-sm text-muted-foreground",
                                children: "Original Request",
                              }),
                              o("div", {
                                className: "mt-1 text-sm font-mono",
                                children: d.originalRequest,
                              }),
                            ],
                          }),
                        ],
                      }),
                    }),
                    o(X, {
                      value: "headers",
                      children: o("div", {
                        className: "bg-gray-50 dark:bg-gray-900 rounded p-4",
                        children: o("div", {
                          className: "space-y-1 text-sm font-mono",
                          children: Object.entries(d.headers).map(([w, _]) =>
                            o(
                              "div",
                              {
                                children: [
                                  o("span", {
                                    className: "text-blue-600",
                                    children: [w, ":"],
                                  }),
                                  " ",
                                  o("span", {
                                    className:
                                      "text-gray-700 dark:text-gray-300",
                                    children: Array.isArray(_)
                                      ? _.join(", ")
                                      : _,
                                  }),
                                ],
                              },
                              w
                            )
                          ),
                        }),
                      }),
                    }),
                    o(X, {
                      value: "body",
                      children: o("div", {
                        className: "bg-gray-50 dark:bg-gray-900 rounded p-4",
                        children: o("pre", {
                          className:
                            "text-sm font-mono overflow-x-auto max-h-96 overflow-y-auto",
                          children: d.body,
                        }),
                      }),
                    }),
                  ],
                }),
              }),
            ],
          }),
      ],
    });
  }
  function tl({ request: e }) {
    let [t, r] = P(new Set()),
      [n, a] = P(null),
      s = Z(
        () =>
          !e.MiddlewareTrace || e.MiddlewareTrace.length === 0
            ? []
            : e.MiddlewareTrace.map((v) => ({
                name: v.name || "Unknown",
                type: v.type || "middleware",
                start_time: v.start_time || "",
                end_time: v.end_time || "",
                duration: v.duration || 0,
                status: v.status || "completed",
                error: v.error,
                details: v.details,
                children: v.children || [],
              })),
        [e.MiddlewareTrace]
      ),
      i = Z(
        () =>
          e.PerformanceMetrics?.sql_queries
            ? e.PerformanceMetrics.sql_queries
            : [],
        [e.PerformanceMetrics]
      ),
      l = Z(
        () =>
          e.PerformanceMetrics?.http_calls
            ? e.PerformanceMetrics.http_calls
            : [],
        [e.PerformanceMetrics]
      ),
      u = (v) => {
        r((y) => {
          let N = new Set(y);
          return N.has(v) ? N.delete(v) : N.add(v), N;
        });
      },
      f = (v) => {
        switch (v) {
          case "middleware":
            return "text-blue-600 bg-blue-50";
          case "handler":
            return "text-green-600 bg-green-50";
          case "sql":
            return "text-purple-600 bg-purple-50";
          case "http":
            return "text-orange-600 bg-orange-50";
          case "custom":
            return "text-gray-600 bg-gray-50";
          default:
            return "text-gray-600 bg-gray-50";
        }
      },
      m = (v) => {
        switch (v) {
          case "completed":
            return "text-green-600";
          case "error":
            return "text-red-600";
          case "running":
            return "text-yellow-600";
          default:
            return "text-gray-600";
        }
      },
      c = (v) =>
        v < 1 ? "<1ms" : v < 1e3 ? `${v}ms` : `${(v / 1e3).toFixed(2)}s`,
      d = (v, y = 0, N = "0") => {
        let R = v.children && v.children.length > 0,
          M = t.has(N);
        return o(
          "div",
          {
            className: "border-l-2 border-gray-200",
            children: [
              o("div", {
                className: D(
                  "flex items-center gap-2 p-2 hover:bg-gray-50 cursor-pointer",
                  y > 0 && "ml-4"
                ),
                onClick: () => {
                  a(v), R && u(N);
                },
                children: [
                  R &&
                    o("span", {
                      className: "text-gray-400 text-sm",
                      children: M ? "\u25BC" : "\u25B6",
                    }),
                  o(ie, {
                    className: D("text-xs", f(v.type)),
                    children: v.type,
                  }),
                  o("span", {
                    className: "flex-1 text-sm font-medium",
                    children: v.name,
                  }),
                  o("span", {
                    className: D("text-xs", m(v.status)),
                    children: v.status,
                  }),
                  o("span", {
                    className: "text-xs text-gray-500",
                    children: c(v.duration),
                  }),
                ],
              }),
              R &&
                M &&
                o("div", {
                  className: "ml-2",
                  children: v.children.map((w, _) => d(w, y + 1, `${N}-${_}`)),
                }),
            ],
          },
          N
        );
      },
      p = (v, y) =>
        o(
          "div",
          {
            className: "border rounded-lg p-3 mb-2",
            children: [
              o("div", {
                className: "flex items-center justify-between mb-2",
                children: [
                  o(ie, {
                    className: "text-xs bg-purple-100 text-purple-800",
                    children: ["SQL Query #", y + 1],
                  }),
                  o("span", {
                    className: "text-xs text-gray-500",
                    children: c(v.duration),
                  }),
                ],
              }),
              o("pre", {
                className:
                  "text-xs bg-gray-50 p-2 rounded overflow-x-auto font-mono",
                children: v.query,
              }),
              o("div", {
                className: "flex items-center gap-4 mt-2 text-xs text-gray-500",
                children: [
                  o("span", { children: ["Rows: ", v.rows || 0] }),
                  v.error &&
                    o("span", {
                      className: "text-red-600",
                      children: ["Error: ", v.error],
                    }),
                ],
              }),
            ],
          },
          y
        ),
      x = (v, y) =>
        o(
          "div",
          {
            className: "border rounded-lg p-3 mb-2",
            children: [
              o("div", {
                className: "flex items-center justify-between mb-2",
                children: [
                  o("div", {
                    className: "flex items-center gap-2",
                    children: [
                      o(ie, {
                        className: "text-xs bg-orange-100 text-orange-800",
                        children: v.method,
                      }),
                      o("span", {
                        className: "text-sm font-medium",
                        children: v.url,
                      }),
                    ],
                  }),
                  o("span", {
                    className: "text-xs text-gray-500",
                    children: c(v.duration),
                  }),
                ],
              }),
              o("div", {
                className: "flex items-center gap-4 text-xs text-gray-500",
                children: [
                  o("span", { children: ["Status: ", v.status] }),
                  o("span", { children: ["Size: ", v.size, " bytes"] }),
                ],
              }),
            ],
          },
          y
        ),
      b = () => {
        if (s.length === 0)
          return o("div", {
            className: "text-center py-8 text-gray-500",
            children: "No trace data available for this request",
          });
        let v = [],
          y = (w, _ = "") => {
            v.push({
              type: "trace",
              name: w.name,
              traceType: w.type,
              startTime: new Date(w.start_time).getTime(),
              endTime: new Date(w.end_time).getTime(),
              duration: w.duration,
              status: w.status,
              parent: _,
            }),
              w.children && w.children.forEach((T) => y(T, w.name));
          };
        if (
          (s.forEach((w) => y(w)),
          v.sort((w, _) => w.startTime - _.startTime),
          v.length === 0)
        )
          return o("div", {
            className: "text-center py-8 text-gray-500",
            children: "No timeline data available",
          });
        let N = v[0].startTime,
          M = Math.max(...v.map((w) => w.endTime || w.startTime)) - N || 1;
        return o("div", {
          className: "relative",
          children: v.map((w, _) => {
            let T = ((w.startTime - N) / M) * 100,
              C = (w.duration / M) * 100;
            return o(
              "div",
              {
                className: "relative h-8 mb-1",
                children: [
                  o("div", {
                    className: "absolute inset-y-0 left-0 w-32 pr-2 text-right",
                    children: o("span", {
                      className: "text-xs truncate",
                      children: w.name,
                    }),
                  }),
                  o("div", {
                    className: "absolute inset-y-0 left-32 right-0",
                    children: o("div", {
                      className: D(
                        "absolute h-6 top-1 rounded",
                        f(w.traceType),
                        w.status === "error" && "border-2 border-red-500"
                      ),
                      style: { left: `${T}%`, width: `${Math.max(C, 1)}%` },
                      title: `${w.name}: ${c(w.duration)}`,
                    }),
                  }),
                ],
              },
              _
            );
          }),
        });
      };
    return o("div", {
      className: "space-y-4",
      children: o(I, {
        children: [
          o(j, { children: o(W, { children: "Request Trace" }) }),
          o(A, {
            children: o(ke, {
              defaultValue: "trace",
              className: "w-full",
              children: [
                o(_e, {
                  className: "grid w-full grid-cols-4",
                  children: [
                    o(Y, { value: "trace", children: "Trace Tree" }),
                    o(Y, { value: "timeline", children: "Timeline" }),
                    o(Y, { value: "sql", children: "SQL Queries" }),
                    o(Y, { value: "http", children: "HTTP Calls" }),
                  ],
                }),
                o(X, {
                  value: "trace",
                  className: "space-y-4",
                  children: [
                    o("div", {
                      className: "border rounded-lg",
                      children:
                        s.length > 0
                          ? s.map((v, y) => d(v, 0, y.toString()))
                          : o("div", {
                              className: "p-8 text-center text-gray-500",
                              children:
                                "No middleware traces recorded for this request",
                            }),
                    }),
                    n &&
                      o(I, {
                        className: "mt-4",
                        children: [
                          o(j, {
                            children: o(W, {
                              className: "text-sm",
                              children: "Trace Details",
                            }),
                          }),
                          o(A, {
                            children: o("div", {
                              className: "space-y-2 text-sm",
                              children: [
                                o("div", {
                                  children: [
                                    o("span", {
                                      className: "font-medium",
                                      children: "Name:",
                                    }),
                                    " ",
                                    n.name,
                                  ],
                                }),
                                o("div", {
                                  children: [
                                    o("span", {
                                      className: "font-medium",
                                      children: "Type:",
                                    }),
                                    " ",
                                    o(ie, {
                                      className: D("text-xs", f(n.type)),
                                      children: n.type,
                                    }),
                                  ],
                                }),
                                o("div", {
                                  children: [
                                    o("span", {
                                      className: "font-medium",
                                      children: "Duration:",
                                    }),
                                    " ",
                                    c(n.duration),
                                  ],
                                }),
                                o("div", {
                                  children: [
                                    o("span", {
                                      className: "font-medium",
                                      children: "Status:",
                                    }),
                                    " ",
                                    o("span", {
                                      className: m(n.status),
                                      children: n.status,
                                    }),
                                  ],
                                }),
                                n.error &&
                                  o("div", {
                                    children: [
                                      o("span", {
                                        className: "font-medium",
                                        children: "Error:",
                                      }),
                                      " ",
                                      o("span", {
                                        className: "text-red-600",
                                        children: n.error,
                                      }),
                                    ],
                                  }),
                                n.details &&
                                  o("div", {
                                    children: [
                                      o("span", {
                                        className: "font-medium",
                                        children: "Details:",
                                      }),
                                      o("pre", {
                                        className:
                                          "mt-2 p-2 bg-gray-50 rounded text-xs overflow-x-auto",
                                        children: JSON.stringify(
                                          n.details,
                                          null,
                                          2
                                        ),
                                      }),
                                    ],
                                  }),
                              ],
                            }),
                          }),
                        ],
                      }),
                  ],
                }),
                o(X, {
                  value: "timeline",
                  children: o("div", {
                    className: "border rounded-lg p-4 overflow-x-auto",
                    children: b(),
                  }),
                }),
                o(X, {
                  value: "sql",
                  children:
                    i.length > 0
                      ? o("div", {
                          children: [
                            o("div", {
                              className:
                                "mb-4 flex items-center justify-between",
                              children: [
                                o("span", {
                                  className: "text-sm text-gray-600",
                                  children: ["Total SQL Queries: ", i.length],
                                }),
                                o("span", {
                                  className: "text-sm text-gray-600",
                                  children: [
                                    "Total Time:",
                                    " ",
                                    c(i.reduce((v, y) => v + y.duration, 0)),
                                  ],
                                }),
                              ],
                            }),
                            i.map((v, y) => p(v, y)),
                          ],
                        })
                      : o("div", {
                          className: "p-8 text-center text-gray-500",
                          children: "No SQL queries recorded for this request",
                        }),
                }),
                o(X, {
                  value: "http",
                  children:
                    l.length > 0
                      ? o("div", {
                          children: [
                            o("div", {
                              className:
                                "mb-4 flex items-center justify-between",
                              children: [
                                o("span", {
                                  className: "text-sm text-gray-600",
                                  children: ["Total HTTP Calls: ", l.length],
                                }),
                                o("span", {
                                  className: "text-sm text-gray-600",
                                  children: [
                                    "Total Time:",
                                    " ",
                                    c(l.reduce((v, y) => v + y.duration, 0)),
                                  ],
                                }),
                              ],
                            }),
                            l.map((v, y) => x(v, y)),
                          ],
                        })
                      : o("div", {
                          className: "p-8 text-center text-gray-500",
                          children: "No HTTP calls recorded for this request",
                        }),
                }),
              ],
            }),
          }),
        ],
      }),
    });
  }
  function rl({ requests: e, onImport: t }) {
    let [r, n] = P(null),
      [a, s] = P(!1),
      i = $(null),
      l = () => {
        try {
          let c = fe.exportRequests(e),
            d = new Blob([c], { type: "application/json" }),
            p = URL.createObjectURL(d),
            x = document.createElement("a");
          (x.href = p),
            (x.download = `govisual-requests-${Date.now()}.json`),
            document.body.appendChild(x),
            x.click(),
            document.body.removeChild(x),
            URL.revokeObjectURL(p);
        } catch (c) {
          console.error("Export failed:", c);
        }
      },
      u = () => {
        try {
          let c = [
              "ID",
              "Timestamp",
              "Method",
              "Path",
              "Status",
              "Duration (ms)",
              "Error",
            ],
            d = e.map((y) => [
              y.ID,
              y.Timestamp,
              y.Method,
              y.Path,
              y.StatusCode,
              y.Duration,
              y.Error || "",
            ]),
            p = [c.join(","), ...d.map((y) => y.map((N) => `"${N}"`).join(","))]
              .join(`
`),
            x = new Blob([p], { type: "text/csv" }),
            b = URL.createObjectURL(x),
            v = document.createElement("a");
          (v.href = b),
            (v.download = `govisual-requests-${Date.now()}.csv`),
            document.body.appendChild(v),
            v.click(),
            document.body.removeChild(v),
            URL.revokeObjectURL(b);
        } catch (c) {
          console.error("CSV export failed:", c);
        }
      },
      f = (c) => {
        let d = c.target,
          p = d.files?.[0];
        if (!p) return;
        n(null), s(!1);
        let x = new FileReader();
        (x.onload = (b) => {
          try {
            let v = b.target?.result,
              y = fe.importRequests(v);
            t(y), s(!0), setTimeout(() => s(!1), 3e3);
          } catch (v) {
            n(v.message || "Failed to import requests"),
              setTimeout(() => n(null), 5e3);
          }
        }),
          x.readAsText(p),
          (d.value = "");
      },
      m = () => {
        i.current?.click();
      };
    return o(I, {
      children: [
        o(j, {
          children: [
            o(W, { children: "Export & Import" }),
            o(Tt, {
              children:
                "Export request logs for analysis or import previously saved logs",
            }),
          ],
        }),
        o(A, {
          className: "space-y-4",
          children: [
            o("div", {
              className: "flex flex-col sm:flex-row gap-4",
              children: [
                o("div", {
                  className: "flex-1",
                  children: [
                    o("h4", {
                      className: "text-sm font-medium mb-2",
                      children: "Export Data",
                    }),
                    o("div", {
                      className: "flex flex-col sm:flex-row gap-2",
                      children: [
                        o(K, {
                          onClick: l,
                          disabled: e.length === 0,
                          className: "flex-1",
                          children: "Export as JSON",
                        }),
                        o(K, {
                          onClick: u,
                          disabled: e.length === 0,
                          variant: "outline",
                          className: "flex-1",
                          children: "Export as CSV",
                        }),
                      ],
                    }),
                    e.length === 0 &&
                      o("p", {
                        className: "text-xs text-muted-foreground mt-2",
                        children: "No requests to export",
                      }),
                    e.length > 0 &&
                      o("p", {
                        className: "text-xs text-muted-foreground mt-2",
                        children: [
                          e.length,
                          " request",
                          e.length !== 1 ? "s" : "",
                          " will be exported",
                        ],
                      }),
                  ],
                }),
                o("div", {
                  className: "flex-1",
                  children: [
                    o("h4", {
                      className: "text-sm font-medium mb-2",
                      children: "Import Data",
                    }),
                    o("input", {
                      ref: i,
                      type: "file",
                      accept: ".json",
                      onChange: f,
                      className: "hidden",
                    }),
                    o(K, {
                      onClick: m,
                      variant: "outline",
                      className: "w-full",
                      children: "Import JSON",
                    }),
                    o("p", {
                      className: "text-xs text-muted-foreground mt-2",
                      children: "Import previously exported request logs",
                    }),
                  ],
                }),
              ],
            }),
            r &&
              o("div", {
                className: "p-3 bg-red-50 border border-red-200 rounded-md",
                children: o("p", {
                  className: "text-sm text-red-800",
                  children: r,
                }),
              }),
            a &&
              o("div", {
                className: "p-3 bg-green-50 border border-green-200 rounded-md",
                children: o("p", {
                  className: "text-sm text-green-800",
                  children: "Requests imported successfully!",
                }),
              }),
          ],
        }),
      ],
    });
  }
  function nl({ requests: e }) {
    let [t, r] = P("24h"),
      [n, a] = P("timeline"),
      s = Z(() => {
        let c = new Date().getTime(),
          d = {
            "1h": 60 * 60 * 1e3,
            "6h": 6 * 60 * 60 * 1e3,
            "24h": 24 * 60 * 60 * 1e3,
            "7d": 7 * 24 * 60 * 60 * 1e3,
          },
          p = c - d[t];
        return e.filter((x) => new Date(x.Timestamp).getTime() > p);
      }, [e, t]),
      i = Z(() => {
        let c = t === "1h" ? 6e4 : t === "6h" ? 3e5 : t === "24h" ? 9e5 : 36e5,
          d = new Map();
        s.forEach((x) => {
          let b = new Date(x.Timestamp).getTime(),
            v = Math.floor(b / c) * c;
          d.has(v) || d.set(v, []), d.get(v).push(x.Duration);
        });
        let p = Array.from(d.entries()).sort((x, b) => x[0] - b[0]);
        return {
          labels: p.map(([x]) => new Date(x).toLocaleTimeString()),
          avgDuration: p.map(
            ([, x]) => x.reduce((b, v) => b + v, 0) / x.length
          ),
          maxDuration: p.map(([, x]) => Math.max(...x)),
          minDuration: p.map(([, x]) => Math.min(...x)),
          count: p.map(([, x]) => x.length),
        };
      }, [s, t]),
      l = Z(() => {
        let c = [0, 100, 200, 500, 1e3, 2e3, 5e3, 1e4, 1 / 0],
          d = [
            "<100ms",
            "100-200ms",
            "200-500ms",
            "500ms-1s",
            "1-2s",
            "2-5s",
            "5-10s",
            ">10s",
          ],
          p = new Array(c.length - 1).fill(0);
        return (
          s.forEach((x) => {
            for (let b = 0; b < c.length - 1; b++)
              if (x.Duration >= c[b] && x.Duration < c[b + 1]) {
                p[b]++;
                break;
              }
          }),
          { labels: d, counts: p }
        );
      }, [s]),
      u = Z(() => {
        let c = new Map();
        return (
          s.forEach((d) => {
            let p = `${d.Method} ${d.Path}`;
            c.has(p) || c.set(p, []), c.get(p).push(d.Duration);
          }),
          Array.from(c.entries())
            .map(([d, p]) => {
              let x = [...p].sort((y, N) => y - N),
                b = Math.floor(x.length * 0.95),
                v = Math.floor(x.length * 0.99);
              return {
                endpoint: d,
                count: p.length,
                avgDuration: p.reduce((y, N) => y + N, 0) / p.length,
                minDuration: Math.min(...p),
                maxDuration: Math.max(...p),
                p95Duration: x[b] || x[x.length - 1],
                p99Duration: x[v] || x[x.length - 1],
              };
            })
            .sort((d, p) => p.avgDuration - d.avgDuration)
        );
      }, [s]),
      f = (c, d, p = 200, x = "#3b82f6") => {
        if (c.length === 0) return null;
        let b = Math.max(...c),
          v = Math.min(...c),
          y = b - v || 1,
          N = 800,
          R = 40,
          M = N - 2 * R,
          w = p - 2 * R,
          _ = M / (c.length - 1 || 1),
          T = w / y,
          C = c.map((U, B) => ({ x: R + B * _, y: R + (b - U) * T })),
          V = C.reduce(
            (U, B, oe) =>
              U + (oe === 0 ? `M ${B.x},${B.y}` : ` L ${B.x},${B.y}`),
            ""
          );
        return o("svg", {
          viewBox: `0 0 ${N} ${p}`,
          className: "w-full h-full",
          children: [
            [0, 1, 2, 3, 4].map((U) => {
              let B = R + (w * U) / 4,
                oe = b - (y * U) / 4;
              return o(
                "g",
                {
                  children: [
                    o("line", {
                      x1: R,
                      y1: B,
                      x2: N - R,
                      y2: B,
                      stroke: "#e5e7eb",
                      strokeWidth: "1",
                    }),
                    o("text", {
                      x: R - 5,
                      y: B + 4,
                      textAnchor: "end",
                      className: "text-xs fill-gray-500",
                      children: [oe.toFixed(0), "ms"],
                    }),
                  ],
                },
                U
              );
            }),
            o("path", { d: V, fill: "none", stroke: x, strokeWidth: "2" }),
            C.map((U, B) =>
              o(
                "circle",
                {
                  cx: U.x,
                  cy: U.y,
                  r: "3",
                  fill: x,
                  className: "hover:r-5 transition-all",
                  children: o("title", {
                    children: `${d[B]}: ${c[B].toFixed(0)}ms`,
                  }),
                },
                B
              )
            ),
          ],
        });
      },
      m = (c, d, p = 200, x = "#3b82f6") => {
        if (c.length === 0) return null;
        let b = Math.max(...c),
          v = 800,
          y = 40,
          N = v - 2 * y,
          R = p - 2 * y,
          M = (N / c.length) * 0.8,
          w = (N / c.length) * 0.2,
          _ = R / (b || 1);
        return o("svg", {
          viewBox: `0 0 ${v} ${p}`,
          className: "w-full h-full",
          children: [
            [0, 1, 2, 3, 4].map((T) => {
              let C = y + (R * T) / 4,
                V = b - (b * T) / 4;
              return o(
                "g",
                {
                  children: [
                    o("line", {
                      x1: y,
                      y1: C,
                      x2: v - y,
                      y2: C,
                      stroke: "#e5e7eb",
                      strokeWidth: "1",
                    }),
                    o("text", {
                      x: y - 5,
                      y: C + 4,
                      textAnchor: "end",
                      className: "text-xs fill-gray-500",
                      children: V.toFixed(0),
                    }),
                  ],
                },
                T
              );
            }),
            c.map((T, C) => {
              let V = y + C * (M + w) + w / 2,
                U = T * _,
                B = y + R - U;
              return o(
                "g",
                {
                  children: [
                    o("rect", {
                      x: V,
                      y: B,
                      width: M,
                      height: U,
                      fill: x,
                      className: "hover:opacity-80 transition-opacity",
                      children: o("title", { children: `${d[C]}: ${T}` }),
                    }),
                    o("text", {
                      x: V + M / 2,
                      y: p - 5,
                      textAnchor: "middle",
                      className: "text-xs fill-gray-500",
                      transform: `rotate(-45, ${V + M / 2}, ${p - 5})`,
                      children: d[C],
                    }),
                  ],
                },
                C
              );
            }),
          ],
        });
      };
    return o(I, {
      children: [
        o(j, {
          children: o("div", {
            className: "flex items-center justify-between",
            children: [
              o(W, { children: "Response Time Analysis" }),
              o("div", {
                className: "flex gap-2",
                children: ["1h", "6h", "24h", "7d"].map((c) =>
                  o(
                    "button",
                    {
                      onClick: () => r(c),
                      className: `px-3 py-1 text-sm rounded ${
                        t === c
                          ? "bg-primary text-primary-foreground"
                          : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                      }`,
                      children: c,
                    },
                    c
                  )
                ),
              }),
            ],
          }),
        }),
        o(A, {
          children:
            s.length === 0
              ? o("div", {
                  className: "text-center py-8 text-muted-foreground",
                  children: "No requests in the selected time range",
                })
              : o(ke, {
                  value: n,
                  onValueChange: (c) => a(c),
                  children: [
                    o(_e, {
                      className: "grid w-full grid-cols-3",
                      children: [
                        o(Y, { value: "timeline", children: "Timeline" }),
                        o(Y, {
                          value: "distribution",
                          children: "Distribution",
                        }),
                        o(Y, { value: "endpoints", children: "Endpoints" }),
                      ],
                    }),
                    o(X, {
                      value: "timeline",
                      className: "space-y-4",
                      children: [
                        o("div", {
                          children: [
                            o("h4", {
                              className: "text-sm font-medium mb-2",
                              children: "Average Response Time",
                            }),
                            o("div", {
                              className: "h-48",
                              children: f(i.avgDuration, i.labels),
                            }),
                          ],
                        }),
                        o("div", {
                          className: "grid grid-cols-2 gap-4",
                          children: [
                            o("div", {
                              children: [
                                o("h4", {
                                  className: "text-sm font-medium mb-2",
                                  children: "Max Response Time",
                                }),
                                o("div", {
                                  className: "h-32",
                                  children: f(
                                    i.maxDuration,
                                    i.labels,
                                    128,
                                    "#ef4444"
                                  ),
                                }),
                              ],
                            }),
                            o("div", {
                              children: [
                                o("h4", {
                                  className: "text-sm font-medium mb-2",
                                  children: "Request Count",
                                }),
                                o("div", {
                                  className: "h-32",
                                  children: f(
                                    i.count,
                                    i.labels,
                                    128,
                                    "#10b981"
                                  ),
                                }),
                              ],
                            }),
                          ],
                        }),
                      ],
                    }),
                    o(X, {
                      value: "distribution",
                      children: o("div", {
                        children: [
                          o("h4", {
                            className: "text-sm font-medium mb-2",
                            children: "Response Time Distribution",
                          }),
                          o("div", {
                            className: "h-64",
                            children: m(l.counts, l.labels, 256, "#8b5cf6"),
                          }),
                        ],
                      }),
                    }),
                    o(X, {
                      value: "endpoints",
                      children: o("div", {
                        className: "overflow-x-auto",
                        children: o("table", {
                          className: "w-full text-sm",
                          children: [
                            o("thead", {
                              children: o("tr", {
                                className: "border-b",
                                children: [
                                  o("th", {
                                    className: "text-left p-2",
                                    children: "Endpoint",
                                  }),
                                  o("th", {
                                    className: "text-right p-2",
                                    children: "Count",
                                  }),
                                  o("th", {
                                    className: "text-right p-2",
                                    children: "Avg",
                                  }),
                                  o("th", {
                                    className: "text-right p-2",
                                    children: "Min",
                                  }),
                                  o("th", {
                                    className: "text-right p-2",
                                    children: "Max",
                                  }),
                                  o("th", {
                                    className: "text-right p-2",
                                    children: "P95",
                                  }),
                                  o("th", {
                                    className: "text-right p-2",
                                    children: "P99",
                                  }),
                                ],
                              }),
                            }),
                            o("tbody", {
                              children: u
                                .slice(0, 10)
                                .map((c) =>
                                  o(
                                    "tr",
                                    {
                                      className: "border-b hover:bg-gray-50",
                                      children: [
                                        o("td", {
                                          className: "p-2 font-mono text-xs",
                                          children: c.endpoint,
                                        }),
                                        o("td", {
                                          className: "text-right p-2",
                                          children: c.count,
                                        }),
                                        o("td", {
                                          className: "text-right p-2",
                                          children: [
                                            c.avgDuration.toFixed(0),
                                            "ms",
                                          ],
                                        }),
                                        o("td", {
                                          className: "text-right p-2",
                                          children: [
                                            c.minDuration.toFixed(0),
                                            "ms",
                                          ],
                                        }),
                                        o("td", {
                                          className: "text-right p-2",
                                          children: [
                                            c.maxDuration.toFixed(0),
                                            "ms",
                                          ],
                                        }),
                                        o("td", {
                                          className: "text-right p-2",
                                          children: [
                                            c.p95Duration.toFixed(0),
                                            "ms",
                                          ],
                                        }),
                                        o("td", {
                                          className: "text-right p-2",
                                          children: [
                                            c.p99Duration.toFixed(0),
                                            "ms",
                                          ],
                                        }),
                                      ],
                                    },
                                    c.endpoint
                                  )
                                ),
                            }),
                          ],
                        }),
                      }),
                    }),
                  ],
                }),
        }),
      ],
    });
  }
  function ol() {
    let [e, t] = P([]),
      [r, n] = P([]),
      [a, s] = P(null),
      [i, l] = P(!1),
      [u, f] = P("dashboard"),
      [m, c] = P({ method: "", statusCode: "", path: "", minDuration: "" }),
      [d, p] = P([]),
      [x, b] = P(!1),
      [v, y] = P(!1),
      [N, R] = P(null);
    q(() => {
      fe.getRequests()
        .then((O) => {
          t(O), n(O);
        })
        .catch(console.error);
      let z = fe.subscribeToEvents((O) => {
        t(O), M(O, m);
      });
      return () => {
        z.close();
      };
    }, []);
    let M = (z, O) => {
        let Q = [...z];
        if (
          (O.method && (Q = Q.filter((E) => E.Method === O.method)),
          O.statusCode)
        ) {
          let E = O.statusCode.charAt(0);
          Q = Q.filter((me) => me.StatusCode.toString().charAt(0) === E);
        }
        if (O.path) {
          let E = O.path.toLowerCase();
          Q = Q.filter((me) => me.Path.toLowerCase().includes(E));
        }
        if (O.minDuration) {
          let E = parseInt(O.minDuration);
          isNaN(E) || (Q = Q.filter((me) => me.Duration >= E));
        }
        n(Q);
      },
      w = (z) => {
        c(z), M(e, z);
      },
      _ = async () => {
        try {
          await fe.clearRequests(), t([]), n([]), s(null), p([]);
        } catch (z) {
          console.error("Failed to clear requests:", z);
        }
      },
      T = (z) => {
        s(z), l(!0);
      },
      C = (z) => {
        p((O) => (O.includes(z) ? O.filter((Q) => Q !== z) : [...O, z]));
      },
      V = () => {
        d.length >= 2 && b(!0);
      },
      U = (z) => {
        R(z), y(!0);
      },
      B = (z) => {
        let O = [...e, ...z],
          Q = Array.from(new Map(O.map((E) => [E.ID, E])).values());
        t(Q), M(Q, m);
      },
      rt = (() => {
        let z = e.length,
          O = e.filter(
            (me) => me.StatusCode >= 200 && me.StatusCode < 300
          ).length,
          Q = z > 0 ? Math.round((O / z) * 100) : 0,
          E =
            z > 0
              ? Math.round(e.reduce((me, xt) => me + xt.Duration, 0) / z)
              : 0;
        return { total: z, successRate: Q, avgDuration: E };
      })();
    return o("div", {
      className: "flex h-screen bg-gradient-to-br from-background to-muted/20",
      children: [
        o(Na, { activeTab: u, onTabChange: f, stats: rt, onClearAll: _ }),
        o("main", {
          className: "flex-1 overflow-y-auto",
          children: o("div", {
            className: "p-8",
            children: (() => {
              switch (u) {
                case "dashboard":
                  return o("div", {
                    className: "space-y-8 animate-in fade-in-50 duration-500",
                    children: [
                      o("div", {
                        className: "mb-2",
                        children: [
                          o("h1", {
                            className: "text-3xl font-bold tracking-tight",
                            children: "Dashboard",
                          }),
                          o("p", {
                            className: "text-muted-foreground mt-1",
                            children:
                              "Monitor and analyze HTTP requests in real-time",
                          }),
                        ],
                      }),
                      o(Ca, { requests: e }),
                      o(Kn, { onFilterChange: w, onClear: _ }),
                      o("div", {
                        className:
                          "bg-background rounded-xl shadow-sm border overflow-hidden",
                        children: [
                          o("div", {
                            className: "px-6 py-4 border-b bg-muted/30",
                            children: o("div", {
                              className: "flex items-center justify-between",
                              children: [
                                o("h3", {
                                  className: "text-lg font-semibold",
                                  children: "Recent Requests",
                                }),
                                o("span", {
                                  className: "text-sm text-muted-foreground",
                                  children: [r.length, " total requests"],
                                }),
                              ],
                            }),
                          }),
                          o("div", {
                            className: "overflow-auto max-h-[500px]",
                            children: o(cr, {
                              requests: r.slice(0, 50),
                              selectedRequest: a,
                              onRequestSelect: T,
                            }),
                          }),
                        ],
                      }),
                    ],
                  });
                case "requests":
                  return o("div", {
                    className: "space-y-8 animate-in fade-in-50 duration-500",
                    children: [
                      o("div", {
                        className: "mb-2",
                        children: [
                          o("h1", {
                            className: "text-3xl font-bold tracking-tight",
                            children: "All Requests",
                          }),
                          o("p", {
                            className: "text-muted-foreground mt-1",
                            children:
                              "View and filter all captured HTTP requests",
                          }),
                        ],
                      }),
                      o(Kn, { onFilterChange: w, onClear: _ }),
                      o("div", {
                        className:
                          "bg-background rounded-xl shadow-sm border overflow-hidden",
                        children: [
                          o("div", {
                            className: "px-6 py-4 border-b bg-muted/30",
                            children: o("div", {
                              className: "flex items-center justify-between",
                              children: [
                                o("h3", {
                                  className: "text-lg font-semibold",
                                  children: "Request Log",
                                }),
                                o("div", {
                                  className: "flex items-center gap-4",
                                  children: [
                                    d.length > 0 &&
                                      o(K, {
                                        onClick: V,
                                        disabled: d.length < 2,
                                        size: "sm",
                                        children: [
                                          "Compare ",
                                          d.length,
                                          " Selected",
                                        ],
                                      }),
                                    o("span", {
                                      className:
                                        "text-sm text-muted-foreground",
                                      children: [
                                        "Showing ",
                                        r.length,
                                        " of ",
                                        e.length,
                                        " ",
                                        "requests",
                                      ],
                                    }),
                                  ],
                                }),
                              ],
                            }),
                          }),
                          o("div", {
                            className: "overflow-auto",
                            style: { height: "calc(100vh - 380px)" },
                            children: o(cr, {
                              requests: r,
                              selectedRequest: a,
                              onRequestSelect: T,
                              selectedForComparison: d,
                              onToggleComparison: C,
                              onReplay: U,
                            }),
                          }),
                        ],
                      }),
                    ],
                  });
                case "environment":
                  return o("div", {
                    className: "space-y-8 animate-in fade-in-50 duration-500",
                    children: [
                      o("div", {
                        className: "mb-2",
                        children: [
                          o("h1", {
                            className: "text-3xl font-bold tracking-tight",
                            children: "Environment",
                          }),
                          o("p", {
                            className: "text-muted-foreground mt-1",
                            children:
                              "System information and environment variables",
                          }),
                        ],
                      }),
                      o(Ji, {}),
                    ],
                  });
                case "trace":
                  return o("div", {
                    className: "space-y-8 animate-in fade-in-50 duration-500",
                    children: [
                      o("div", {
                        className: "mb-2",
                        children: [
                          o("h1", {
                            className: "text-3xl font-bold tracking-tight",
                            children: "Request Trace",
                          }),
                          o("p", {
                            className: "text-muted-foreground mt-1",
                            children:
                              "Analyze request execution flow, middleware chain, SQL queries, and HTTP calls",
                          }),
                        ],
                      }),
                      a
                        ? o(tl, { request: a })
                        : o("div", {
                            children: [
                              o("div", {
                                className:
                                  "mb-4 p-4 bg-blue-50 border border-blue-200 rounded-lg",
                                children: o("p", {
                                  className: "text-sm text-blue-800",
                                  children:
                                    "Select a request from the table below to view its execution trace, including middleware execution, SQL queries, and external HTTP calls.",
                                }),
                              }),
                              o("div", {
                                className:
                                  "bg-background rounded-xl shadow-sm border overflow-hidden",
                                children: [
                                  o("div", {
                                    className: "px-6 py-4 border-b bg-muted/30",
                                    children: o("h3", {
                                      className: "text-lg font-semibold",
                                      children: "Select a Request to Trace",
                                    }),
                                  }),
                                  o("div", {
                                    className: "overflow-auto",
                                    style: { height: "calc(100vh - 380px)" },
                                    children: o(cr, {
                                      requests: r.slice(0, 100),
                                      selectedRequest: a,
                                      onRequestSelect: T,
                                      selectedForComparison: d,
                                      onToggleComparison: C,
                                      onReplay: U,
                                    }),
                                  }),
                                ],
                              }),
                            ],
                          }),
                    ],
                  });
                case "analytics":
                  return o("div", {
                    className: "space-y-8 animate-in fade-in-50 duration-500",
                    children: [
                      o("div", {
                        className: "mb-2",
                        children: [
                          o("h1", {
                            className: "text-3xl font-bold tracking-tight",
                            children: "Analytics",
                          }),
                          o("p", {
                            className: "text-muted-foreground mt-1",
                            children:
                              "Performance metrics and request analysis",
                          }),
                        ],
                      }),
                      o(nl, { requests: e }),
                      o(rl, { requests: r, onImport: B }),
                    ],
                  });
                default:
                  return null;
              }
            })(),
          }),
        }),
        o(Ki, { request: a, open: i, onOpenChange: l }),
        x &&
          o("div", {
            className:
              "fixed inset-0 bg-black/50 flex items-center justify-center z-50",
            children: o("div", {
              className:
                "bg-background rounded-lg p-6 max-w-7xl max-h-[90vh] overflow-auto",
              children: o(Zi, {
                requestIds: d,
                allRequests: e,
                onClose: () => {
                  b(!1), p([]);
                },
              }),
            }),
          }),
        v &&
          N &&
          o("div", {
            className:
              "fixed inset-0 bg-black/50 flex items-center justify-center z-50",
            children: o("div", {
              className:
                "bg-background rounded-lg p-6 max-w-4xl max-h-[90vh] overflow-auto",
              children: o(el, {
                request: N,
                onClose: () => {
                  y(!1), R(null);
                },
              }),
            }),
          }),
      ],
    });
  }
  var al = document.getElementById("app");
  al ? Oe(o(ol, {}), al) : console.error("Could not find app root element");
})();
