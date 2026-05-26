import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { Hexagon, Cpu, Link as LinkIcon, ArrowRight, Activity, Globe, BarChart, FileText, Search, Brain, Network } from 'lucide-react';

export default function LandingPage() {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-background text-text overflow-hidden font-sans">
      {/* Navbar Container */}
      <nav className="fixed w-full z-50 border-b border-slate-700/50 bg-background/80 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Hexagon className="w-8 h-8 text-primary" strokeWidth={1.5} />
            <span className="font-bold text-xl tracking-wide text-white">
              TrustChain <span className="text-primary">AI</span>
            </span>
          </div>
          <div className="hidden md:flex items-center space-x-8 text-sm font-medium">
            <a href="#features" className="text-textMuted hover:text-white transition">Fitur</a>
            <a href="#about" className="text-textMuted hover:text-white transition">Platform</a>
            <a href="#compliance" className="text-textMuted hover:text-white transition">Kepatuhan</a>
          </div>
          <div className="flex items-center space-x-4">
            <button 
              onClick={() => navigate('/login')}
              className="text-white text-sm font-medium hover:text-primary transition"
            >
              Masuk
            </button>
            <button 
              onClick={() => navigate('/login')}
              className="px-5 py-2.5 bg-primary/10 border border-primary/30 text-primary rounded-lg text-sm font-semibold hover:bg-primary hover:text-white transition-all shadow-[0_0_15px_rgba(14,165,233,0.3)]"
            >
              Mulai Sekarang
            </button>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative pt-32 pb-20 md:pt-48 md:pb-32 px-6">
        <div className="absolute top-1/4 left-1/4 w-[500px] h-[500px] bg-primary/20 rounded-full blur-[150px] mix-blend-screen pointer-events-none"></div>
        <div className="absolute bottom-1/4 right-1/4 w-[600px] h-[600px] bg-accent/10 rounded-full blur-[150px] mix-blend-screen pointer-events-none"></div>

        <div className="max-w-7xl mx-auto relative z-10 text-center">
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
          >
            <span className="inline-flex items-center space-x-2 px-4 py-1.5 rounded-full bg-surface border border-slate-700/50 text-sm font-medium text-slate-300 mb-6">
              <span className="flex h-2 w-2 relative">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-accent"></span>
              </span>
              <span>Deteksi Fraud Perusahaan v2.4 Aktif</span>
            </span>
            
            <h1 className="text-5xl md:text-7xl font-extrabold text-white tracking-tight mb-8 leading-tight">
              Mengamankan Masa Depan <br className="hidden md:block"/>
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-primary to-accent">Rantai Pasok & Keuangan Energi</span>
            </h1>
            
            <p className="text-lg md:text-xl text-slate-400 max-w-2xl mx-auto mb-10 leading-relaxed">
              Deteksi anomali secara real-time dengan AI mutakhir. Pastikan audibilitas absolut melalui buku besar blockchain yang tidak dapat diubah. Dirancang khusus untuk alur kerja industri yang kompleks.
            </p>
            
            <div className="flex flex-col sm:flex-row items-center justify-center space-y-4 sm:space-y-0 sm:space-x-6">
              <button 
                onClick={() => navigate('/login')}
                className="w-full sm:w-auto px-8 py-4 bg-primary text-white rounded-xl font-bold transition-all shadow-[0_0_25px_rgba(14,165,233,0.4)] hover:shadow-[0_0_40px_rgba(14,165,233,0.6)] hover:bg-primary/90 flex items-center justify-center"
              >
                Masuk Platform <ArrowRight className="ml-2 w-5 h-5" />
              </button>
              <button className="w-full sm:w-auto px-8 py-4 bg-surface hover:bg-surfaceHover border border-slate-700 rounded-xl text-white font-bold transition flex items-center justify-center">
                <Globe className="mr-2 w-5 h-5 text-slate-400" /> Lihat Arsitektur
              </button>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Features Grid */}
      <section id="features" className="py-24 bg-surface/30 backdrop-blur-xl border-y border-slate-800/50">
        <div className="max-w-7xl mx-auto px-6">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">Pendekatan Mesin Ganda</h2>
            <p className="text-slate-400 max-w-2xl mx-auto">Menggabungkan Kecerdasan Buatan prediktif dengan kepastian kriptografi dari buku besar terdistribusi.</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {[
              {
                icon: Cpu,
                title: "F1: Deteksi Fraud AI Real-Time",
                desc: "Model Isolation Forest + LSTM mendeteksi anomali secara real-time. Adaptive learning memperbarui model otomatis dengan false positive yang rendah.",
                color: "text-primary",
                bg: "bg-primary/10"
              },
              {
                icon: LinkIcon,
                title: "F2: Audit Immutable Blockchain",
                desc: "Transaksi di-hash di Hyperledger/Ethereum. Smart Contract memvalidasi otomatis, memungkinkan verifikasi data pihak ketiga secara independen.",
                color: "text-accent",
                bg: "bg-accent/10"
              },
              {
                icon: BarChart,
                title: "F3: Skor Risiko Dinamis",
                desc: "Skor risiko 0-100 untuk transaksi/vendor. Visualisasi heatmap dan analisis tren untuk identifikasi pola risiko sebelum menjadi fraud.",
                color: "text-warning",
                bg: "bg-warning/10"
              },
              {
                icon: FileText,
                title: "F4: Kepatuhan Otomatis",
                desc: "Hasilkan laporan kepatuhan otomatis sesuai OJK & BI (PBI/POJK). Ekspor ke PDF/Excel, terintegrasi ERP (SAP/Oracle) via REST API.",
                color: "text-primary",
                bg: "bg-primary/10"
              },
              {
                icon: Search,
                title: "F5: Manajemen Kasus & Investigasi",
                desc: "Manajemen kasus fraud end-to-end. Fitur kolaborasi tim, anotasi, tugas, dan timeline yang terhubung langsung ke bukti blockchain.",
                color: "text-danger",
                bg: "bg-danger/10"
              },
              {
                icon: Brain,
                title: "F6: Modul Explainable AI",
                desc: "Dapatkan wawasan transparan atas keputusan AI. Pahami alasan di balik setiap anomali yang dideteksi untuk memastikan akuntabilitas.",
                color: "text-accent",
                bg: "bg-accent/10"
              },
              {
                icon: Network,
                title: "F7: Mesin Federated Learning",
                desc: "Latih model AI secara kolaboratif antar berbagai institusi tanpa perlu membagikan data mentah yang sensitif (privasi terjaga penuh).",
                color: "text-warning",
                bg: "bg-warning/10"
              }
            ].map((feature, idx) => (
              <motion.div 
                key={idx}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: idx * 0.2 }}
                className="bg-slate-800/40 border border-slate-700/50 rounded-2xl p-8 hover:bg-slate-800/60 transition duration-300"
              >
                <div className={`w-14 h-14 rounded-xl ${feature.bg} flex items-center justify-center mb-6`}>
                  <feature.icon className={`w-7 h-7 ${feature.color}`} />
                </div>
                <h3 className="text-xl font-bold text-white mb-3">{feature.title}</h3>
                <p className="text-slate-400 leading-relaxed">{feature.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* Integrations Banner */}
      <section className="py-20 px-6">
        <div className="max-w-7xl mx-auto text-center">
          <p className="text-sm font-semibold tracking-widest text-slate-500 uppercase mb-8">Integrasi mulus dengan berbagai ERP utama</p>
          <div className="flex flex-wrap justify-center items-center gap-12 opacity-60 grayscale hover:grayscale-0 transition-all duration-500">
            <div className="font-bold text-2xl text-white flex items-center"><Activity className="mr-2" /> SAP S/4HANA</div>
            <div className="font-bold text-2xl text-white">ORACLE NetSuite</div>
            <div className="font-bold text-2xl text-white">Microsoft Dynamics 365</div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-slate-800 bg-surface text-sm py-12 px-6">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row justify-between items-center">
          <div className="flex items-center space-x-2 mb-4 md:mb-0">
            <Hexagon className="w-5 h-5 text-primary" />
            <span className="font-bold text-white tracking-wide">TrustChain AI</span>
          </div>
          <p className="text-slate-500">© 2026 TrustChain Technologies. Hak Cipta Dilindungi.</p>
        </div>
      </footer>
    </div>
  );
}
